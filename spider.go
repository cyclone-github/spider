package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

/*
   cyclone's spider
   spider will crawl a url and create a wordlist, or use flag -ngram to create ngrams
v0.5.10;
	initial github release
v0.6.2;
   fixed scraping logic & ngram creations bugs
   switched from gocolly to goquery for web scraping
   remove dups from word / ngrams output
v0.7.0;
   added feature to allow crawling specific file extensions (html, htm, txt)
   added check to keep crawler from crawling offsite URLs
   added flag "-delay" to avoid rate limiting (-delay 100 == 100ms delay between URL requests)
   added write buffer for better performance on large files
   increased crawl depth from 5 to 100 (not recommended, but enabled for edge cases)
   fixed out of bounds slice bug when crawling URLs with NIL characters
   fixed bug when attempting to crawl deeper than available URLs to crawl
   fixed crawl depth calculation
   optimized code which runs 2.8x faster vs v0.6.x during bench testing
v0.7.1;
    added progress bars to word / ngrams processing & file writing operations
    added RAM usage monitoring
    optimized order of operations for faster processing with less RAM
	TO-DO: refactor code (func main is getting messy)
    TO-DO: add -file flag to allow crawling local plaintext files such as an ebook.txt (COMPLETED in v0.8.0)
v0.8.0;
    added flag "-file" to allow creating ngrams from a local plaintext file (ex: foobar.txt)
    added flag "-timeout" for -url mode
    added flag "-sort" which sorts output by frequency
    fixed several small bugs
v0.8.1;
	updated default -delay to 10ms
v0.9.0;
	added flag "-url-match" to only crawl URLs containing a specified keyword; https://github.com/cyclone-github/spider/issues/6
	added notice to user if no URLs are crawled when using "-crawl 1 -url-match"
	exit early if zero URLs were crawled (no processing or file output)
	use custom User-Agent "Spider/0.9.0 (+https://github.com/cyclone-github/spider)"
	removed clearScreen function and its imports
	fixed crawl-depth calculation logic
	fixed restrict link collection to .html, .htm, .txt and extension-less paths
	upgraded dependencies and bumped Go version to v1.24.3
v0.9.1;
	added flag "-agent" to allow user to specify custom user-agent; https://github.com/cyclone-github/spider/issues/8
v1.0.0;
	added flag "-text-match" to filter page text matches
	memory and performance optimizations for -file and -url modes
	-file mode streams wordlists from disk instead of loading entire files into RAM
	reduced RAM usage for large -sort wordlists
	default -timeout increased from 1 to 10 seconds
	progress bars, stats, and errors now write to stderr
	sanitize url fragments for dedup and extension checks
	updated default User-Agent
*/

const spiderVersion = "1.0.0"

// approved link extensions for crawling
var validLinkSuffixes = map[string]bool{
	".html": true,
	".htm":  true,
	".txt":  true,
}

// goquery
func getDocumentFromURL(targetURL string, timeout time.Duration, agent string) (*goquery.Document, bool, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", agent)

	res, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, false, nil
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	return doc, true, err
}

// strip url fragment for dedup and fetching
func sanitizeURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

// check link path extension (.html, .htm, .txt, or none)
func isAllowedLink(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	ext := strings.ToLower(path.Ext(parsed.Path))
	return ext == "" || validLinkSuffixes[ext]
}

func getLinksFromDocument(doc *goquery.Document, baseURL string) []string {
	var links []string

	doc.Find("a[href]").Each(func(_ int, item *goquery.Selection) {
		href, exists := item.Attr("href")
		if !exists || strings.HasPrefix(href, "#") {
			return
		}
		absoluteLink := joinURL(baseURL, href)
		if absoluteLink == "" {
			return
		}

		cleanLink, err := sanitizeURL(absoluteLink)
		if err != nil {
			return
		}

		// only allow approved extensions or none at all
		if !isAllowedLink(cleanLink) {
			return
		}
		links = append(links, cleanLink)
	})
	return links
}

func getTextFromDocument(doc *goquery.Document) string {
	doc.Find("script, style").Each(func(_ int, item *goquery.Selection) {
		item.Remove()
	})
	return doc.Text()
}

func crawlAndScrape(u string, depth, delay int, timeout time.Duration, agent string, fetchedChan, matchedChan chan<- int, textsChan chan<- string, visited map[string]bool, urlMatchStr, textMatchStr string) {
	cleanURL, err := sanitizeURL(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing URL %s: %v\n", u, err)
		return
	}
	u = cleanURL

	if visited[u] {
		return
	}
	visited[u] = true // mark before fetch to avoid retry on error

	doc, isSuccess, err := getDocumentFromURL(u, timeout, agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL %s: %v\n", u, err)
		return
	}
	if !isSuccess {
		return
	}

	fetchedChan <- 1 // count every successful page fetch

	// only scrape text if -url-match and -text-match pass (links still crawled either way)
	if urlMatchStr == "" || strings.Contains(strings.ToLower(u), urlMatchStr) {
		text := getTextFromDocument(doc)
		if textMatchStr == "" || strings.Contains(strings.ToLower(text), textMatchStr) {
			matchedChan <- 1  // page included in wordlist
			textsChan <- text // send the text for later n-gram processing
		}
	}

	if depth > 1 {
		baseDomain, err := getBaseDomain(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting base domain: %v\n", err)
			return
		}
		for _, link := range getLinksFromDocument(doc, u) {
			time.Sleep(time.Duration(delay) * time.Millisecond)

			linkDomain, err := getBaseDomain(link)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing link %s: %v\n", link, err)
				continue
			}
			if linkDomain != baseDomain {
				continue
			}

			// only *descend* into children that match (if urlMatchStr was provided)
			if urlMatchStr != "" && !strings.Contains(strings.ToLower(link), urlMatchStr) {
				continue
			}

			crawlAndScrape(link, depth-1, delay, timeout, agent, fetchedChan, matchedChan, textsChan, visited, urlMatchStr, textMatchStr)
		}
	}
}

func getBaseDomain(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return parsedURL.Hostname(), nil
}

// joinURL function to handle relative URLs
func joinURL(baseURL, relativeURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	newURL, err := u.Parse(relativeURL)
	if err != nil {
		return ""
	}
	return newURL.String()
}

// live crawl status on stderr (scan/match only when -text-match is set)
func printCrawlProgress(fetched, matched int, textMatchMode, final bool) {
	if textMatchMode {
		if final {
			fmt.Fprintf(os.Stderr, "\rScan/match:\t%d/%d\n", fetched, matched)
			return
		}
		fmt.Fprintf(os.Stderr, "\rScan/match:\t%d/%d", fetched, matched)
		return
	}
	if final {
		fmt.Fprintf(os.Stderr, "\rURLs scanned:\t%d\n", fetched)
		return
	}
	fmt.Fprintf(os.Stderr, "\rURLs scanned:\t%d", fetched)
}

func updateProgressBar(action string, total, processed int) {
	if total == 0 {
		return // avoid division by zero
	}
	percentage := float64(processed) / float64(total) * 100
	fmt.Fprintf(os.Stderr, "\r%s...\t[", action)
	for i := 0; i < int(percentage/5); i++ {
		fmt.Fprint(os.Stderr, "=")
	}
	for i := int(percentage / 5); i < 20; i++ {
		fmt.Fprint(os.Stderr, " ")
	}
	fmt.Fprintf(os.Stderr, "] %.2f%%", percentage)
}

// track bytes read from file for progress bar updates
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func markUniqueWords(words []string, uniqueWords map[string]bool) {
	for _, word := range words {
		uniqueWords[word] = true // count unique words
	}
}

// build n-grams from a word slice (url mode and in-memory text)
func countNgramsFromWords(words []string, ngramMin, ngramMax int, ngramCounts map[string]int) {
	for i := 0; i <= len(words)-ngramMin; i++ {
		for n := ngramMin; n <= ngramMax && i+n <= len(words); n++ {
			ngramCounts[strings.Join(words[i:i+n], " ")]++ // count n-gram frequency
		}
	}
}

// process scraped page text into unique words and n-gram counts
func processTextBlob(text string, ngramMin, ngramMax int, uniqueWords map[string]bool, ngramCounts map[string]int, trackUnique bool) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}
	if trackUnique {
		markUniqueWords(words, uniqueWords)
	}
	if ngramMax == 1 {
		for _, word := range words {
			ngramCounts[word]++ // count word frequency
		}
		return
	}
	countNgramsFromWords(words, ngramMin, ngramMax, ngramCounts)
}

// stream file from disk instead of loading entire file into ram
func countNgramsFromStream(r io.Reader, fileSize int64, ngramMin, ngramMax int, uniqueWords map[string]bool, ngramCounts map[string]int, trackUnique bool, progress func(processed, total int)) error {
	cr := &countingReader{r: r}
	scanner := bufio.NewScanner(cr)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(bufio.ScanWords)

	total := int(fileSize)
	if total == 0 {
		total = 1
	}

	if ngramMax == 1 {
		for scanner.Scan() {
			word := scanner.Text()
			if trackUnique {
				uniqueWords[word] = true
			}
			ngramCounts[word]++ // count word frequency
			if progress != nil {
				progress(int(cr.n), total)
			}
		}
		return scanner.Err()
	}

	// sliding window for multi-word n-grams while streaming
	window := make([]string, 0, ngramMax)
	for scanner.Scan() {
		word := scanner.Text()
		if trackUnique {
			uniqueWords[word] = true
		}
		window = append(window, word)
		for n := ngramMin; n <= ngramMax && n <= len(window); n++ {
			start := len(window) - n
			ngramCounts[strings.Join(window[start:], " ")]++ // count n-gram frequency
		}
		if len(window) > ngramMax {
			window = window[1:]
		}
		if progress != nil {
			progress(int(cr.n), total)
		}
	}
	return scanner.Err()
}

func monitorRAMUsage(stopChan chan bool, maxRAMUsage *float64) {
	var memStats runtime.MemStats
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&memStats)
			currentUsage := float64(memStats.Alloc) / 1024 / 1024 / 1024 // GB
			if currentUsage > *maxRAMUsage {
				*maxRAMUsage = currentUsage
			}
		case <-stopChan:
			return
		}
	}
}

// write unique n-grams to output file
func writeNgrams(outPath string, ngramCounts map[string]int, sortByFreq bool) error {
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriterSize(outFile, 1*1024*1024) // 1MB buffer
	totalNgrams := len(ngramCounts)
	interval := totalNgrams / 100
	if interval == 0 {
		interval = 1
	}

	if sortByFreq {
		fmt.Fprintln(os.Stderr, "Sorting wordlist by frequency...")
		type pair struct {
			Text  string
			Count int
		}
		pairs := make([]pair, 0, len(ngramCounts)) // preallocate sort slice
		for txt, cnt := range ngramCounts {
			pairs = append(pairs, pair{txt, cnt})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].Count != pairs[j].Count {
				return pairs[i].Count > pairs[j].Count
			}
			return pairs[i].Text < pairs[j].Text
		})
		for i, p := range pairs {
			if _, err := writer.WriteString(p.Text + "\n"); err != nil {
				return err
			}
			if i%interval == 0 {
				updateProgressBar("Writing", len(pairs), i+1)
			}
		}
	} else {
		// original unsorted output
		i := 0
		for gram := range ngramCounts {
			if _, err := writer.WriteString(gram + "\n"); err != nil {
				return err
			}
			if i%interval == 0 {
				updateProgressBar("Writing", totalNgrams, i+1)
			}
			i++
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}
	updateProgressBar("Writing", totalNgrams, totalNgrams)
	return nil
}

// main function
func main() {
	cycloneFlag := flag.Bool("cyclone", false, "Display coded message")
	versionFlag := flag.Bool("version", false, "Display version")
	urlFlag := flag.String("url", "", "URL of the website to scrape")
	fileFlag := flag.String("file", "", "Path to a local file to scrape")
	ngramFlag := flag.String("ngram", "1", "Lengths of n-grams (e.g., \"1-3\" for 1, 2, and 3-length n-grams).")
	oFlag := flag.String("o", "", "Output file for the n-grams")
	crawlFlag := flag.Int("crawl", 1, "Depth of links to crawl")
	delayFlag := flag.Int("delay", 10, "Delay in ms between each URL lookup to avoid rate limiting")
	timeoutFlag := flag.Int("timeout", 10, "Timeout for URL crawling in seconds")
	sortFlag := flag.Bool("sort", false, "Sort output by frequency")
	urlMatchFlag := flag.String("url-match", "", "Only crawl URLs containing this keyword (case-insensitive)")
	textMatchFlag := flag.String("text-match", "", "Only process page text containing this keyword (case-insensitive); all URLs are still crawled")
	agentFlag := flag.String("agent", "Mozilla/5.0 (X11) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36 Spider/"+spiderVersion+"", "Custom user-agent")

	flag.Parse()

	if *cycloneFlag {
		codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
		decoded, _ := base64.StdEncoding.DecodeString(codedBy)
		fmt.Fprintln(os.Stderr, string(decoded))
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Fprintf(os.Stderr, "Cyclone's Spider v%s\n", spiderVersion)
		os.Exit(0)
	}

	// sanity check for -url or -file
	if (*urlFlag == "" && *fileFlag == "") || (*urlFlag != "" && *fileFlag != "") {
		fmt.Fprintln(os.Stderr, "Error: You must specify either -url or -file, but not both")
		fmt.Fprintln(os.Stderr, "Try running -help for more information")
		os.Exit(1)
	}
	fileMode := *fileFlag != ""

	urlMatchStr := strings.ToLower(*urlMatchFlag)
	textMatchStr := strings.ToLower(*textMatchFlag)

	if fileMode && textMatchStr != "" {
		fmt.Fprintln(os.Stderr, "Error: -text-match is only supported with -url")
		os.Exit(1)
	}

	var baseDomain string
	if !fileMode {
		// URL mode
		if *crawlFlag < 1 || *crawlFlag > 100 {
			fmt.Fprintln(os.Stderr, "Error: -crawl flag must be between 1 and 100")
			os.Exit(1)
		}
		if *delayFlag < 0 || *delayFlag > 60000 {
			fmt.Fprintln(os.Stderr, "Error: -delay flag must be between 0 and 60000")
			os.Exit(1)
		}
		if *timeoutFlag < 1 || *timeoutFlag > 600 {
			fmt.Fprintln(os.Stderr, "Error: -timeout flag must be between 1 and 600")
			os.Exit(1)
		}

		// check for "http*" on urlFlag so goquery doesn't wet the bed
		u, err := url.Parse(*urlFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing URL: %v\n", err)
			os.Exit(1)
		}
		if u.Scheme == "" {
			u.Scheme = "https"
		}
		cleanURL, err := sanitizeURL(u.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing URL: %v\n", err)
			os.Exit(1)
		}
		*urlFlag = cleanURL

		baseDomain, err = getBaseDomain(*urlFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting base domain: %v\n", err)
			os.Exit(1)
		}
	}

	// parse ngram range
	ngramRange := strings.Split(*ngramFlag, "-")
	ngramMin, err := strconv.Atoi(ngramRange[0])
	if err != nil || ngramMin < 1 || ngramMin > 20 {
		fmt.Fprintln(os.Stderr, "Error: -ngram flag must be between 1 and 20")
		os.Exit(1)
	}
	ngramMax := ngramMin
	if len(ngramRange) > 1 {
		ngramMax, err = strconv.Atoi(ngramRange[1])
		if err != nil || ngramMax < ngramMin || ngramMax > 20 {
			fmt.Fprintln(os.Stderr, "Error: -ngram flag must be between 1 and 20")
			os.Exit(1)
		}
	}

	// default output filename
	if *oFlag == "" {
		if fileMode {
			base := filepath.Base(*fileFlag)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			*oFlag = name + "_spider.txt"
		} else {
			parsedURL, _ := url.Parse(*urlFlag)
			*oFlag = strings.TrimPrefix(parsedURL.Hostname(), "www.") + "_spider.txt"
		}
	}

	timeoutDur := time.Duration(*timeoutFlag) * time.Second
	start := time.Now()

	fmt.Fprintln(os.Stderr, " ------------------ ")
	fmt.Fprintln(os.Stderr, "| Cyclone's Spider |")
	fmt.Fprintln(os.Stderr, " ------------------ ")
	fmt.Fprintln(os.Stderr)
	if fileMode {
		fmt.Fprintf(os.Stderr, "Reading file:\t%s\n", *fileFlag)
		fmt.Fprintf(os.Stderr, "ngram len:\t%s\n", *ngramFlag)
	} else {
		fmt.Fprintf(os.Stderr, "Crawling URL:\t%s\n", *urlFlag)
		fmt.Fprintf(os.Stderr, "Base domain:\t%s\n", baseDomain)
		fmt.Fprintf(os.Stderr, "Crawl depth:\t%d\n", *crawlFlag)
		fmt.Fprintf(os.Stderr, "ngram len:\t%s\n", *ngramFlag)
		fmt.Fprintf(os.Stderr, "Crawl delay:\t%dms (increase to avoid rate limiting)\n", *delayFlag)
		fmt.Fprintf(os.Stderr, "Timeout:\t%d sec\n", *timeoutFlag)
		if urlMatchStr != "" {
			fmt.Fprintf(os.Stderr, "URL match:\t%s\n", *urlMatchFlag)
		}
		if textMatchStr != "" {
			fmt.Fprintf(os.Stderr, "Text match:\t%s\n", *textMatchFlag)
		}
	}

	// start RAM usage monitor
	stopMonitor := make(chan bool)
	var maxRAMUsage float64
	go monitorRAMUsage(stopMonitor, &maxRAMUsage)

	// skip redundant uniqueWordsMap when ngram=1 (use len(ngramCounts) instead)
	trackUnique := !(ngramMin == 1 && ngramMax == 1)
	ngramCounts := make(map[string]int)
	var uniqueWordsMap map[string]bool
	if trackUnique {
		uniqueWordsMap = make(map[string]bool)
	}

	// set up progress bar ticker
	progressTicker := time.NewTicker(100 * time.Millisecond) // update progress every 100ms
	defer progressTicker.Stop()

	// mode-specific input and processing
	if fileMode {
		// read file from disk instead of crawling
		inFile, err := os.Open(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", *fileFlag, err)
			os.Exit(1)
		}
		defer inFile.Close()

		fi, err := inFile.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", *fileFlag, err)
			os.Exit(1)
		}

		lastProgress := 0
		progressFn := func(processed, total int) {
			select {
			case <-progressTicker.C:
				if processed != lastProgress {
					updateProgressBar("Processing", total, processed)
					lastProgress = processed
				}
			default:
			}
		}

		if err := countNgramsFromStream(inFile, fi.Size(), ngramMin, ngramMax, uniqueWordsMap, ngramCounts, trackUnique, progressFn); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", *fileFlag, err)
			os.Exit(1)
		}
		updateProgressBar("Processing", int(fi.Size()), int(fi.Size()))
	} else {
		// URL mode: crawl
		fetchedChan := make(chan int)
		matchedChan := make(chan int)
		textsChan := make(chan string, 64) // buffered channel for text
		doneChan := make(chan struct{})
		var wg sync.WaitGroup

		textMatchMode := textMatchStr != ""

		// live crawl counter on stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
			totalFetched := 0
			totalMatched := 0
			for {
				select {
				case <-ticker.C:
					printCrawlProgress(totalFetched, totalMatched, textMatchMode, false)
				case <-fetchedChan:
					totalFetched++
				case <-matchedChan:
					totalMatched++
				case <-doneChan:
					ticker.Stop()
					printCrawlProgress(totalFetched, totalMatched, textMatchMode, true)
					return
				}
			}
		}()

		// start crawling process in goroutine
		visitedURLs := make(map[string]bool)
		wg.Add(1)
		go func() {
			defer wg.Done()
			crawlAndScrape(*urlFlag, *crawlFlag, *delayFlag, timeoutDur, *agentFlag, fetchedChan, matchedChan, textsChan, visitedURLs, urlMatchStr, textMatchStr)
			time.Sleep(100 * time.Millisecond)
			close(textsChan)
			close(doneChan)
		}()

		// collect all text into a slice
		var texts []string
		for text := range textsChan {
			texts = append(texts, text)
		}
		wg.Wait()

		// if nothing matched, exit early
		if len(texts) == 0 {
			fmt.Fprintln(os.Stderr, "No matching page text found, exiting...") // boo, something went wrong!
			if *crawlFlag == 1 {
				fmt.Fprintln(os.Stderr, "Try increasing -crawl depth, or remove -url-match / -text-match")
			}
			stopMonitor <- true
			os.Exit(1)
		}

		// process scraped page text into n-grams
		for _, text := range texts {
			processTextBlob(text, ngramMin, ngramMax, uniqueWordsMap, ngramCounts, trackUnique)
		}
	}

	if len(ngramCounts) == 0 {
		if fileMode {
			fmt.Fprintln(os.Stderr, "No words found, exiting...")
		} else {
			fmt.Fprintln(os.Stderr, "No matching page text found, exiting...")
			if *crawlFlag == 1 {
				fmt.Fprintln(os.Stderr, "Try increasing -crawl depth, or remove -url-match / -text-match")
			}
		}
		stopMonitor <- true
		os.Exit(1)
	}

	// stats
	uniqueWordCount := len(ngramCounts)
	if trackUnique {
		uniqueWordCount = len(uniqueWordsMap)
	}

	if fileMode {
		fmt.Fprintf(os.Stderr, "\nUnique words:\t%d\n", uniqueWordCount)
	} else {
		fmt.Fprintf(os.Stderr, "Unique words:\t%d\n", uniqueWordCount)
	}
	fmt.Fprintf(os.Stderr, "Unique ngrams:\t%d\n", len(ngramCounts))

	// write unique n-grams to file
	if err := writeNgrams(*oFlag, ngramCounts, *sortFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		stopMonitor <- true
		os.Exit(1)
	}

	stopMonitor <- true

	// print statistics
	fmt.Fprintf(os.Stderr, "\nOutput file:\t%s\n", *oFlag)
	fmt.Fprintf(os.Stderr, "RAM used:\t%.3f GB\n", maxRAMUsage)
	fmt.Fprintf(os.Stderr, "Runtime:\t%.3fs\n", time.Since(start).Seconds())
}

// end code
