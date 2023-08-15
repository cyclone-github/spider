package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// cyclone's url spider
// spider will crawl a url and create a wordlist, or use flag -ngram to create ngrams
// version 0.5.10; initial github release
/* version 0.6.2;
   fixed scraping logic & ngram creations bugs
   switched from gocolly to goquery for web scraping
   remove dups from word / ngrams output
*/

// clear screen function
func clearScreen() {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// goquery
func getDocumentFromURL(targetURL string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return goquery.NewDocumentFromReader(res.Body)
}

func getLinksFromDocument(doc *goquery.Document) []string {
	var links []string
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		links = append(links, link)
	})
	return links
}

func getTextFromDocument(doc *goquery.Document) string {
	doc.Find("script, style").Each(func(index int, item *goquery.Selection) {
		item.Remove()
	})
	return doc.Text()
}

func crawlAndScrape(u string, depth int, phrase int) map[string]bool {
	ngrams := make(map[string]bool)
	doc, err := getDocumentFromURL(u)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return ngrams
	}
	text := getTextFromDocument(doc)
	for _, ngram := range generateNgrams(text, phrase) {
		ngrams[ngram] = true
	}

	if depth > 1 {
		links := getLinksFromDocument(doc)
		for _, link := range links[:depth-1] {
			absoluteLink := joinURL(u, link)
			childNgrams := crawlAndScrape(absoluteLink, depth-1, phrase)
			for ngram := range childNgrams {
				ngrams[ngram] = true
			}
		}
	}

	return ngrams
}

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

func generateNgrams(text string, n int) []string {
	words := strings.Fields(text)
	var ngrams []string
	for i := 0; i < len(words)-n+1; i++ {
		ngrams = append(ngrams, strings.Join(words[i:i+n], " "))
	}
	return ngrams
}

func uniqueStrings(str string) map[string]bool {
	words := strings.Fields(str)
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		uniqueWords[word] = true
	}
	return uniqueWords
}

func uniqueStringsSlice(strs []string) map[string]bool {
	uniqueStrings := make(map[string]bool)
	for _, str := range strs {
		uniqueStrings[str] = true
	}
	return uniqueStrings
}

// main function
func main() {
	clearScreen()

	cycloneFlag := flag.Bool("cyclone", false, "Display coded message")
	versionFlag := flag.Bool("version", false, "Display version")
	urlFlag := flag.String("url", "", "URL of the website to scrape")
	ngramFlag := flag.String("ngram", "1", "Lengths of n-grams (e.g., \"1-3\" for 1, 2, and 3-length n-grams). Default: 1")
	oFlag := flag.String("o", "", "Output file for the n-grams")
	crawlFlag := flag.Int("crawl", 1, "Number of links to crawl (default: 1)")
	flag.Parse()

	if *cycloneFlag {
		codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
		codedByDecoded, _ := base64.StdEncoding.DecodeString(codedBy)
		fmt.Fprintln(os.Stderr, string(codedByDecoded))
		os.Exit(0)
	}

	if *versionFlag {
		version := "Q3ljbG9uZSdzIFVSTCBTcGlkZXIgdjAuNi4yCg=="
		versionDecoded, _ := base64.StdEncoding.DecodeString(version)
		fmt.Fprintln(os.Stderr, string(versionDecoded))
		os.Exit(0)
	}

	if *urlFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -url flag is required")
		fmt.Fprintln(os.Stderr, "Try running --help for more information")
		os.Exit(1)
	}

	if *crawlFlag < 1 || *crawlFlag > 5 {
		fmt.Fprintln(os.Stderr, "Error: -crawl flag must be between 1 and 5")
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
	fmt.Fprintf(os.Stderr, "Crawl depth:\t%d\n", *crawlFlag)
	fmt.Fprintf(os.Stderr, "ngram len:\t%s\n", *ngramFlag)

	ngrams := make(map[string]bool)
	for i := ngramMin; i <= ngramMax; i++ {
		for ngram := range crawlAndScrape(*urlFlag, *crawlFlag, i) {
			ngrams[ngram] = true
		}
	}

	// extract n-grams into a slice
	var ngramSlice []string
	for ngram := range ngrams {
		ngramSlice = append(ngramSlice, ngram)
	}

	// write unique n-grams to file
	file, err := os.Create(*oFlag)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	for _, ngram := range ngramSlice {
		file.WriteString(ngram + "\n")
	}

	// calculate unique words
	uniqueWords := len(uniqueStrings(strings.Join(ngramSlice, " ")))

	// calculate unique n-grams
	uniqueNgrams := len(ngramSlice)

	runtime := time.Since(start)

	// print statistics
	fmt.Fprintf(os.Stderr, "Unique words:\t%d\n", uniqueWords)
	fmt.Fprintf(os.Stderr, "Unique ngrams:\t%d\n", uniqueNgrams)
	fmt.Fprintf(os.Stderr, "Saved to:\t%s\n", *oFlag)
	fmt.Fprintf(os.Stderr, "Runtime:\t%.6fs\n", runtime.Seconds())
}

// end code
