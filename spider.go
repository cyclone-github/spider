package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// cyclone's url spider
// version 0.5.10; initial github release

// global variables... I know...
var (
	urlFlag     string
	crawlFlag   int
	oFlag       string
	phraseFlag  int
	cycloneFlag bool
	versionFlag bool
	wordList    = make(map[string]int)
	wordListMu  sync.Mutex
)

// initilize flags
func init() {
	flag.StringVar(&urlFlag, "url", "", "URL to scrape")
	flag.IntVar(&crawlFlag, "crawl", 1, "Depth to crawl links")
	flag.StringVar(&oFlag, "o", "", "Output file for word list")
	flag.IntVar(&phraseFlag, "phrase", 1, "Process pairs of words")
	flag.BoolVar(&cycloneFlag, "cyclone", false, "")
	flag.BoolVar(&versionFlag, "version", false, "Version number")
	flag.Parse()

	// check for "http*" on urlFlag so gocolly doesn't wet the bed
	u, err := url.Parse(urlFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing URL: %v\n", err)
		os.Exit(1)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
		urlFlag = u.String()
	}
}

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

// word processing logic
func processWords(text string, phrase int) {
	// acquire lock before accessing wordList
	wordListMu.Lock()
	defer wordListMu.Unlock()

	wordRegex := regexp.MustCompile(`\w+`)
	words := wordRegex.FindAllString(text, -1)

	for i := 0; i < len(words); i++ {
		if i+phrase <= len(words) {
			phraseWords := make([]string, phrase)
			for j := 0; j < phrase; j++ {
				phraseWords[j] = words[i+j]
			}
			phraseStr := strings.Join(phraseWords, " ")
			if _, ok := wordList[phraseStr]; ok {
				wordList[phraseStr]++
			} else {
				wordList[phraseStr] = 1
			}
		} else {
			word := words[i]
			if _, ok := wordList[word]; ok {
				wordList[word]++
			} else {
				wordList[word] = 1
			}
		}
	}
}

// save wordlist logic
func saveWordList(filename string) {
	// acquire lock before accessing wordList
	wordListMu.Lock()
	defer wordListMu.Unlock()
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	type wordCount struct {
		Word  string
		Count int
	}

	var counts []wordCount
	for word, count := range wordList {
		counts = append(counts, wordCount{Word: word, Count: count})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].Count > counts[j].Count
	})

	for _, wc := range counts {
		_, err := fmt.Fprintln(file, wc.Word)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	}
}

// main function
func main() {
	clearScreen()

	if cycloneFlag {
		codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
		codedByDecoded, _ := base64.StdEncoding.DecodeString(codedBy)
		fmt.Fprintln(os.Stderr, string(codedByDecoded))
		os.Exit(0)
	}

	if versionFlag {
		version := "Q3ljbG9uZSdzIFVSTCBTcGlkZXIgdjAuNS4xMAo="
		versionDecoded, _ := base64.StdEncoding.DecodeString(version)
		fmt.Fprintln(os.Stderr, string(versionDecoded))
		os.Exit(0)
	}

	if urlFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -url flag is required")
		os.Exit(1)
	}

	if crawlFlag < 1 || crawlFlag > 100 {
		fmt.Fprintln(os.Stderr, "Error: -crawl flag must be between 1 and 100")
		os.Exit(1)
	}

	if phraseFlag < 1 || phraseFlag > 100 {
		fmt.Fprintln(os.Stderr, "Error: -phrase flag must be between 1 and 100")
		os.Exit(1)
	}

	if oFlag == "" {
		parsedUrl, err := url.Parse(urlFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing URL")
			os.Exit(1)
		}
		// default wordlist output if -oFlag is not specified
		oFlag = strings.TrimPrefix(parsedUrl.Hostname(), "www.") + "_wordlist.txt"
	}

	start := time.Now()

	fmt.Fprintln(os.Stderr, " ---------------------- ")
	fmt.Fprintln(os.Stderr, "| Cyclone's URL Spider |")
	fmt.Fprintln(os.Stderr, " ---------------------- ")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Crawling URL:\t%s\n", urlFlag)

	c := colly.NewCollector(
		colly.MaxDepth(crawlFlag),
		colly.Async(true),
	)

	// initialize depth to crawlFlag
	depth := crawlFlag

	// print crawl & depth info
	fmt.Fprintf(os.Stderr, "Crawl Depth:\t%d\n", depth)
	fmt.Fprintf(os.Stderr, "Phrase Depth:\t%d\n", phraseFlag)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" && depth > 0 { // check if depth is greater than 0
			depth-- // decrement depth after visiting a link
			e.Request.Visit(link)
			time.Sleep(250 * time.Millisecond) // add short sleep time between requests to keep from being rate limited
		}
	})

	// only collect text from these elements using colly.HTML
	c.OnHTML("p, h1, h2, h3, h4, h5, h6, li", func(e *colly.HTMLElement) {
		processWords(e.Text, phraseFlag)
	})

	c.OnScraped(func(r *colly.Response) {
		saveWordList(oFlag)
	})

	err := c.Visit(urlFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error crawling URL: %v\n", err)
		os.Exit(1)
	}

	c.Wait()

	// print runtime results
	fmt.Fprintf(os.Stderr, "Unique words:\t%d\n", len(wordList))
	fmt.Fprintf(os.Stderr, "Wordlist:\t%s\n", oFlag)
	fmt.Fprintf(os.Stderr, "Runtime:\t%s\n", time.Since(start))
}

// end code
