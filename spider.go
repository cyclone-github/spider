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
	"runtime"
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
func getDocumentFromURL(targetURL string) (*goquery.Document, bool, error) {
	client := &http.Client{}
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
		linkTag := item
		link, exists := linkTag.Attr("href")
		if exists {
			absoluteLink := joinURL(baseURL, link) // convert to absolute URL
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

func crawlAndScrape(u string, depth int, delay int, urlCountChan chan<- int, textsChan chan<- string, visited map[string]bool) {
	if visited[u] {
		return
	}

	doc, isSuccess, err := getDocumentFromURL(u)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	if !isSuccess {
		return
	}

	visited[u] = true
	urlCountChan <- 1 // URL processed

	text := getTextFromDocument(doc)
	textsChan <- text // send the text for later n-gram processing

	baseDomain, err := getBaseDomain(u)
	if err != nil {
		fmt.Println("Error getting base domain:", err)
		return
	}

	if depth > 1 {
		links := getLinksFromDocument(doc, u)
		for _, link := range links {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			absoluteLink := joinURL(u, link)
			linkDomain, err := getBaseDomain(absoluteLink)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting link domain for %s: %v\n", absoluteLink, err)
				continue
			}
			if linkDomain == baseDomain {
				crawlAndScrape(absoluteLink, depth-1, delay, urlCountChan, textsChan, visited)
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
	ngramFlag := flag.String("ngram", "1", "Lengths of n-grams (e.g., \"1-3\" for 1, 2, and 3-length n-grams).")
	oFlag := flag.String("o", "", "Output file for the n-grams")
	crawlFlag := flag.Int("crawl", 1, "Depth of links to crawl")
	delayFlag := flag.Int("delay", 0, "Delay in ms between each URL lookup to avoid rate limiting")
	flag.Parse()

	if *cycloneFlag {
		codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
		codedByDecoded, _ := base64.StdEncoding.DecodeString(codedBy)
		fmt.Fprintln(os.Stderr, string(codedByDecoded))
		os.Exit(0)
	}

	if *versionFlag {
		version := "Q3ljbG9uZSdzIFVSTCBTcGlkZXIgdjAuNy4xLWJldGEK"
		versionDecoded, _ := base64.StdEncoding.DecodeString(version)
		fmt.Fprintln(os.Stderr, string(versionDecoded))
		os.Exit(0)
	}

	if *urlFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -url flag is required")
		fmt.Fprintln(os.Stderr, "Try running -help for more information")
		os.Exit(1)
	}

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

	baseDomain, err := getBaseDomain(*urlFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting base domain: %v\n", err)
		os.Exit(1)
	}
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

	if *oFlag == "" {
		parsedUrl, err := url.Parse(*urlFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing URL")
			os.Exit(1)
		}
		*oFlag = strings.TrimPrefix(parsedUrl.Hostname(), "www.") + "_wordlist.txt"
	}

	start := time.Now()

	fmt.Fprintln(os.Stderr, " ---------------------- ")
	fmt.Fprintln(os.Stderr, "| Cyclone's URL Spider |")
	fmt.Fprintln(os.Stderr, " ---------------------- ")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Crawling URL:\t%s\n", *urlFlag)
	fmt.Fprintf(os.Stderr, "Base domain:\t%s\n", baseDomain)
	fmt.Fprintf(os.Stderr, "Crawl depth:\t%d\n", *crawlFlag)
	fmt.Fprintf(os.Stderr, "ngram len:\t%s\n", *ngramFlag)
	fmt.Fprintf(os.Stderr, "Crawl delay:\t%dms (increase this to avoid rate limiting, ex: -delay 100)\n", *delayFlag)

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

	// goroutine to print URLs crawled
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(50 * time.Millisecond)
		totalCrawled := 1
		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\rURLs crawled:\t%d", totalCrawled)
			case count := <-urlCountChan:
				totalCrawled += count
			case <-doneChan:
				fmt.Fprintf(os.Stderr, "\rURLs crawled:\t%d", totalCrawled) // final update
				return
			}
		}
	}()

	// start crawling process in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		crawlAndScrape(*urlFlag, *crawlFlag, *delayFlag, urlCountChan, textsChan, visitedURLs)
		time.Sleep(100 * time.Millisecond)
		close(textsChan)
		close(doneChan)
		fmt.Println()
	}()

	// initialize maps for unique word and n-gram counting
	uniqueWordsMap := make(map[string]bool)
	uniqueNgramsMap := make(map[string]bool)

	// collect all texts into a slice
	var texts []string
	for text := range textsChan {
		texts = append(texts, text)
	}
	totalTexts := len(texts)

	// set up progress bar ticker
	progressTicker := time.NewTicker(100 * time.Millisecond) // update progress every 100ms
	defer progressTicker.Stop()
	processedTexts := 0

	// process texts and generate n-grams
	for _, text := range texts {
		words := strings.Fields(text)
		for _, word := range words {
			uniqueWordsMap[word] = true // count unique words
		}

		for i := 0; i <= len(words)-ngramMin; i++ {
			for n := ngramMin; n <= ngramMax && i+n <= len(words); n++ {
				ngram := strings.Join(words[i:i+n], " ")
				uniqueNgramsMap[ngram] = true // count unique n-grams
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

	// convert unique n-grams map back to a slice for writing to file
	var ngramSlice []string
	for ngram := range uniqueNgramsMap {
		ngramSlice = append(ngramSlice, ngram)
	}

	// calculated counts
	uniqueWords := len(uniqueWordsMap)
	uniqueNgrams := len(uniqueNgramsMap)
	fmt.Fprintf(os.Stderr, "\nUnique words:\t%d\n", uniqueWords)
	fmt.Fprintf(os.Stderr, "Unique ngrams:\t%d\n", uniqueNgrams)

	// write unique n-grams to file
	file, err := os.Create(*oFlag)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 1*1024*1024) // 1MB buffer for better write performance
	totalNgrams := len(ngramSlice)

	// progress update interval
	progressUpdateInterval := totalNgrams / 100
	if progressUpdateInterval == 0 {
		progressUpdateInterval = 1
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for i, ngram := range ngramSlice {
		_, err := writer.WriteString(ngram + "\n")
		if err != nil {
			fmt.Println("Error writing to buffer:", err)
			return
		}
		if i%progressUpdateInterval == 0 {
			updateProgressBar("Writing", totalNgrams, i+1) // update write progress bar
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer to file:", err)
		return
	}
	updateProgressBar("Writing", totalNgrams, totalNgrams) // final update to write progress bar

	// stop RAM monitoring
	stopMonitor <- true

	// print statistics
	fmt.Fprintf(os.Stderr, "\nOutput file:\t%s\n", *oFlag)
	fmt.Fprintf(os.Stderr, "RAM used:\t%.2f GB\n", maxRAMUsage)
	runTime := time.Since(start)
	fmt.Fprintf(os.Stderr, "Runtime:\t%.3fs\n", runTime.Seconds())
}

// end code