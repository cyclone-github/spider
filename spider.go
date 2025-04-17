package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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
   cyclone's url spider
   spider will crawl a url and create a wordlist, or use flag -ngram to create ngrams
version 0.5.10; initial github release
version 0.6.2;
   fixed scraping logic & ngram creations bugs
   switched from gocolly to goquery for web scraping
   remove dups from word / ngrams output
version 0.7.0;
   added feature to allow crawling specific file extensions (html, htm, txt)
   added check to keep crawler from crawling offsite URLs
   added flag "-delay" to avoid rate limiting (-delay 100 == 100ms delay between URL requests)
   added write buffer for better performance on large files
   increased crawl depth from 5 to 100 (not recommended, but enabled for edge cases)
   fixed out of bounds slice bug when crawling URLs with NIL characters
   fixed bug when attempting to crawl deeper than available URLs to crawl
   fixed crawl depth calculation
   optimized code which runs 2.8x faster vs v0.6.x during bench testing
version 0.7.1;
    added progress bars to word / ngrams processing & file writing operations
    added RAM usage monitoring
    optimized order of operations for faster processing with less RAM
    TO-DO: refactor code (func main is getting messy)
    TO-DO: add -file flag to allow crawling local plaintext files such as an ebook.txt
v0.8.0;
    added flag "-file" to allow creating ngrams from a local plaintext file (ex: foobar.txt)
    added flag "-timeout" for -url mode
    added flag "-sort" which sorts output by frequency
    fixed several small bugs
*/

// clear screen function
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		fmt.Fprintln(os.Stderr, "Unsupported platform")
		os.Exit(1)
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clear screen: %v\n", err)
		os.Exit(1)
	}
}

// goquery
func getDocumentFromURL(targetURL string, timeout time.Duration) (*goquery.Document, bool, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

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

func hasAnySuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

func getLinksFromDocument(doc *goquery.Document, baseURL string) []string {
	var links []string
	validSuffixes := []string{".html", ".htm", ".txt"} // specifically crawl file types, ex: if listed in a file server

	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		link, exists := item.Attr("href")
		if exists {
			absoluteLink := joinURL(baseURL, link) // convert to absolute URL
			// crawl any non-anchor or valid-file-type link
			if hasAnySuffix(link, validSuffixes) || !strings.HasPrefix(link, "#") {
				links = append(links, absoluteLink)
			}
		}
	})
	return links
}

func getTextFromDocument(doc *goquery.Document) string {
	doc.Find("script, style").Each(func(index int, item *goquery.Selection) {
		item.Remove()
	})
	return doc.Text()
}

func crawlAndScrape(u string, depth int, delay int, timeout time.Duration, urlCountChan chan<- int, textsChan chan<- string, visited map[string]bool) {
	if visited[u] {
		return
	}
	visited[u] = true // mark before fetch to avoid retry on error

	doc, isSuccess, err := getDocumentFromURL(u, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL %s: %v\n", u, err)
		return
	}
	if !isSuccess {
		return
	}
	urlCountChan <- 1 // URL processed

	text := getTextFromDocument(doc)
	textsChan <- text // send the text for later n-gram processing

	if depth > 1 {
		baseDomain, err := getBaseDomain(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting base domain: %v\n", err)
			return
		}
		links := getLinksFromDocument(doc, u)
		for _, link := range links {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			linkDomain, err := getBaseDomain(link)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing link %s: %v\n", link, err)
				continue
			}
			if linkDomain == baseDomain {
				crawlAndScrape(link, depth-1, delay, timeout, urlCountChan, textsChan, visited)
			}
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

func updateProgressBar(action string, total, processed int) {
	if total == 0 {
		return // avoid division by zero
	}
	percentage := float64(processed) / float64(total) * 100
	fmt.Printf("\r%s...\t[", action)
	for i := 0; i < int(percentage/5); i++ {
		fmt.Print("=")
	}
	for i := int(percentage / 5); i < 20; i++ {
		fmt.Print(" ")
	}
	fmt.Printf("] %.2f%%", percentage)
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

// main function
func main() {
	clearScreen()

	cycloneFlag := flag.Bool("cyclone", false, "Display coded message")
	versionFlag := flag.Bool("version", false, "Display version")
	urlFlag := flag.String("url", "", "URL of the website to scrape")
	fileFlag := flag.String("file", "", "Path to a local file to scrape")
	ngramFlag := flag.String("ngram", "1", "Lengths of n-grams (e.g., \"1-3\" for 1, 2, and 3-length n-grams).")
	oFlag := flag.String("o", "", "Output file for the n-grams")
	crawlFlag := flag.Int("crawl", 1, "Depth of links to crawl")
	delayFlag := flag.Int("delay", 0, "Delay in ms between each URL lookup to avoid rate limiting")
	timeoutFlag := flag.Int("timeout", 1, "Timeout for URL crawling in seconds")
	sortFlag := flag.Bool("sort", false, "Sort output by frequency")
	flag.Parse()

	if *cycloneFlag {
		codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
		decoded, _ := base64.StdEncoding.DecodeString(codedBy)
		fmt.Fprintln(os.Stderr, string(decoded))
		os.Exit(0)
	}
	if *versionFlag {
		version := "Cyclone's URL Spider v0.8.0"
		fmt.Fprintln(os.Stderr, version)
		os.Exit(0)
	}

	// sanity check for -url or -file
	if (*urlFlag == "" && *fileFlag == "") || (*urlFlag != "" && *fileFlag != "") {
		fmt.Fprintln(os.Stderr, "Error: You must specify either -url or -file, but not both")
		fmt.Fprintln(os.Stderr, "Try running -help for more information")
		os.Exit(1)
	}
	fileMode := *fileFlag != ""

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

		// check for "http*" on urlFlag so goquery doesn't wet the bed
		u, err := url.Parse(*urlFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing URL: %v\n", err)
			os.Exit(1)
		}
		if u.Scheme == "" {
			u.Scheme = "https"
			*urlFlag = u.String()
		}

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
			parsedUrl, _ := url.Parse(*urlFlag)
			*oFlag = strings.TrimPrefix(parsedUrl.Hostname(), "www.") + "_spider.txt"
		}
	}

	timeoutDur := time.Duration(*timeoutFlag) * time.Second

	start := time.Now()

	fmt.Fprintln(os.Stderr, " ---------------------- ")
	fmt.Fprintln(os.Stderr, "| Cyclone's URL Spider |")
	fmt.Fprintln(os.Stderr, " ---------------------- ")
	fmt.Fprintln(os.Stderr)
	if fileMode {
		fmt.Fprintf(os.Stderr, "Reading file:\t%s\n", *fileFlag)
		fmt.Fprintf(os.Stderr, "ngram len:\t%s\n", *ngramFlag)
	} else {
		fmt.Fprintf(os.Stderr, "Crawling URL:\t%s\n", *urlFlag)
		fmt.Fprintf(os.Stderr, "Base domain:\t%s\n", baseDomain)
		fmt.Fprintf(os.Stderr, "Crawl depth:\t%d\n", *crawlFlag)
		fmt.Fprintf(os.Stderr, "ngram len:\t%s\n", *ngramFlag)
		fmt.Fprintf(os.Stderr, "Crawl delay:\t%dms (increase this to avoid rate limiting)\n", *delayFlag)
		fmt.Fprintf(os.Stderr, "Timeout:\t%d sec\n", *timeoutFlag)
	}

	// initialize channels and sync group
	urlCountChan := make(chan int)
	textsChan := make(chan string, 1*1024*1024) // buffered channel for text
	visitedURLs := make(map[string]bool)
	doneChan := make(chan struct{})
	var wg sync.WaitGroup
	stopMonitor := make(chan bool)
	var maxRAMUsage float64

	// start RAM usage monitor
	go monitorRAMUsage(stopMonitor, &maxRAMUsage)

	// mode-specific input
	if fileMode {
		// read whole file instead of crawling
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := os.ReadFile(*fileFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", *fileFlag, err)
				os.Exit(1)
			}
			textsChan <- string(data)
			close(textsChan)
		}()
	} else {
		// URL mode: crawl
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
			totalCrawled := 1
			for {
				select {
				case <-ticker.C:
					fmt.Fprintf(os.Stderr, "\rURLs crawled:\t%d", totalCrawled)
				case count := <-urlCountChan:
					totalCrawled += count
				case <-doneChan:
					fmt.Fprintf(os.Stderr, "\rURLs crawled:\t%d", totalCrawled)
					return
				}
			}
		}()

		// start crawling process in goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			crawlAndScrape(*urlFlag, *crawlFlag, *delayFlag, timeoutDur, urlCountChan, textsChan, visitedURLs)
			time.Sleep(100 * time.Millisecond)
			close(textsChan)
			close(doneChan)
			fmt.Println()
		}()
	}

	// collect all text into a slice
	var texts []string
	for text := range textsChan {
		texts = append(texts, text)
	}
	totalTexts := len(texts)

	// set up progress bar ticker
	progressTicker := time.NewTicker(100 * time.Millisecond) // update progress every 100ms
	defer progressTicker.Stop()
	processedTexts := 0

	// maps for unique words and n-gram counts
	uniqueWordsMap := make(map[string]bool)
	ngramCounts := make(map[string]int)

	// process texts and generate n-grams
	for _, text := range texts {
		words := strings.Fields(text)
		for _, word := range words {
			uniqueWordsMap[word] = true // count unique words
		}
		for i := 0; i <= len(words)-ngramMin; i++ {
			for n := ngramMin; n <= ngramMax && i+n <= len(words); n++ {
				ngram := strings.Join(words[i:i+n], " ")
				ngramCounts[ngram]++ // count n-gram frequency
			}
		}
		processedTexts++
		select {
		case <-progressTicker.C:
			updateProgressBar("Processing", totalTexts, processedTexts)
		default:
			// continue without blocking if ticker channel is not ready
		}
	}
	// final update to progress bar output
	updateProgressBar("Processing", totalTexts, processedTexts)

	// stats
	fmt.Fprintf(os.Stderr, "\nUnique words:\t%d\n", len(uniqueWordsMap))
	fmt.Fprintf(os.Stderr, "Unique ngrams:\t%d\n", len(ngramCounts))

	// write unique n-grams to file
	outFile, err := os.Create(*oFlag)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()
	writer := bufio.NewWriterSize(outFile, 1*1024*1024) // 1MB buffer
	totalNgrams := len(ngramCounts)
	interval := totalNgrams / 100
	if interval == 0 {
		interval = 1
	}

	if *sortFlag {
		fmt.Fprintln(os.Stderr, "Sorting n-grams by frequency...")
		type pair struct {
			Text  string
			Count int
		}
		var pairs []pair
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
			_, err := writer.WriteString(p.Text + "\n")
			if err != nil {
				fmt.Println("Error writing to buffer:", err)
				return
			}
			if i%interval == 0 {
				updateProgressBar("Writing", len(pairs), i+1)
			}
		}

	} else {
		// original unsorted output
		i := 0
		for gram := range ngramCounts {
			_, err := writer.WriteString(gram + "\n")
			if err != nil {
				fmt.Println("Error writing to buffer:", err)
				return
			}
			if i%interval == 0 {
				updateProgressBar("Writing", totalNgrams, i+1)
			}
			i++
		}
	}

	if err := writer.Flush(); err != nil {
		fmt.Println("Error flushing buffer to file:", err)
		return
	}
	updateProgressBar("Writing", totalNgrams, totalNgrams)

	// stop RAM monitoring
	stopMonitor <- true

	// print statistics
	fmt.Fprintf(os.Stderr, "\nOutput file:\t%s\n", *oFlag)
	fmt.Fprintf(os.Stderr, "RAM used:\t%.2f GB\n", maxRAMUsage)
	fmt.Fprintf(os.Stderr, "Runtime:\t%.3fs\n", time.Since(start).Seconds())
}

// end code
