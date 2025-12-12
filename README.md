[![Readme Card](https://github-readme-stats-fast.vercel.app/api/pin/?username=cyclone-github&repo=spider&theme=gruvbox)](https://github.com/cyclone-github/spider)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/spider)](https://goreportcard.com/report/github.com/cyclone-github/spider)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/issues)
[![License](https://img.shields.io/github/license/cyclone-github/spider.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/spider.svg)](https://pkg.go.dev/github.com/cyclone-github/spider)

Spider is a web crawler and wordlist/ngram generator written in Go that crawls specified URLs or local files to produce frequency-sorted wordlists and ngrams. Users can customize crawl depth, output files, frequency sort, and ngram options, making it ideal for web scraping to create targeted wordlists for tools like hashcat or John the Ripper. Since Spider is written in Go, it requires no additional libraries to download or install.
*Spider just works*.

# Spider: URL Mode
```
spider -url 'https://forum.hashpwn.net' -crawl 2 -delay 20 -sort -ngram 1-3 -timeout 1 -url-match wordlist -o forum.hashpwn.net_spider.txt -agent 'foobar agent'
```
```
 ---------------------- 
| Cyclone's URL Spider |
 ---------------------- 

Crawling URL:   https://forum.hashpwn.net
Base domain:    forum.hashpwn.net
Crawl depth:    2
ngram len:      1-3
Crawl delay:    20ms (increase this to avoid rate limiting)
Timeout:        1 sec
URLs crawled:   2
Processing...   [====================] 100.00%
Unique words:   475
Unique ngrams:  1977
Sorting n-grams by frequency...
Writing...      [====================] 100.00%
Output file:    forum.hashpwn.net_spider.txt
RAM used:       0.02 GB
Runtime:        2.283s
```
# Spider: File Mode
```
spider -file kjv_bible.txt -sort -ngram 1-3
```
```
 ---------------------- 
| Cyclone's URL Spider |
 ---------------------- 

Reading file:   kjv_bible.txt
ngram len:      1-3
Processing...   [====================] 100.00%
Unique words:   35412
Unique ngrams:  877394
Sorting n-grams by frequency...
Writing...      [====================] 100.00%
Output file:    kjv_bible_spider.txt
RAM used:       0.13 GB
Runtime:        1.359s
```

Wordlist & ngram creation tool to crawl a given url or process a local file to create wordlists and/or ngrams (depending on flags given).
### Usage Instructions:
- To create a simple wordlist from a specified url (will save deduplicated wordlist to url_spider.txt):
  - `spider -url 'https://github.com/cyclone-github'`
- To set url crawl url depth of 2 and create ngrams len 1-5, use flag "-crawl 2" and "-ngram 1-5"
  - `spider -url 'https://github.com/cyclone-github' -crawl 2 -ngram 1-5`
- To set a custom output file, use flag "-o filename"
  - `spider -url 'https://github.com/cyclone-github' -o wordlist.txt`
- To set a delay to keep from being rate-limited, use flag "-delay nth" where nth is time in milliseconds
  - `spider -url 'https://github.com/cyclone-github' -delay 100`
- To set a URL timeout, use flag "-timeout nth" where nth is time in seconds
  - `spider -url 'https://github.com/cyclone-github' -timeout 2`
- To create ngrams len 1-3 and sort output by frequency, use "-ngram 1-3" "-sort"
  - `spider -url 'https://github.com/cyclone-github' -ngram 1-3 -sort`
- To filter crawled URLs by keyword "foobar"
  - `spider -url 'https://github.com/cyclone-github' -url-match foobar`
- To specify a custom user-agent
  - `spider -url 'https://github.com/cyclone-github' -agent 'foobar'`
- To process a local text file, create ngrams len 1-3 and sort output by frequency
  - `spider -file foobar.txt -ngram 1-3 -sort`
- Run `spider -help` to see a list of all options

### spider -help
```
  -agent string
        Custom user-agent (default "Spider/0.9.1 (+https://github.com/cyclone-github/spider)")
  -crawl int
        Depth of links to crawl (default 1)
  -cyclone
        Display coded message
  -delay int
        Delay in ms between each URL lookup to avoid rate limiting (default 10)
  -file string
        Path to a local file to scrape
  -ngram string
        Lengths of n-grams (e.g., "1-3" for 1, 2, and 3-length n-grams). (default "1")
  -o string
        Output file for the n-grams
  -sort
        Sort output by frequency
  -timeout int
        Timeout for URL crawling in seconds (default 1)
  -url string
        URL of the website to scrape
  -url-match string
        Only crawl URLs containing this keyword (case-insensitive)
  -version
        Display version
```
### Install latest release:
```
go install github.com/cyclone-github/spider@latest
```
### Install from latest source code (bleeding edge):
```
go install github.com/cyclone-github/spider@main
```
### Compile from source:
- If you want the latest features, compiling from source is the best option since the release version may run several revisions behind the source code.
- This assumes you have Go and Git installed
  - `git clone https://github.com/cyclone-github/spider.git`   # clone repo
  - `cd spider`                                                # enter project directory
  - `go mod init spider`                                       # initialize Go module (skips if go.mod exists)
  - `go mod tidy`                                              # download dependencies
  - `go build -ldflags="-s -w" .`                              # compile binary in current directory
  - `go install -ldflags="-s -w" .`                            # compile binary and install to $GOPATH
- Compile from source code how-to:
  - https://github.com/cyclone-github/scripts/blob/main/intro_to_go.txt
### Changelog:
- https://github.com/cyclone-github/spider/blob/main/CHANGELOG.md
### Mentions:
- Go Package Documentation: https://pkg.go.dev/github.com/cyclone-github/spider
- Softpedia: https://www.softpedia.com/get/Internet/Other-Internet-Related/Cyclone-s-URL-Spider.shtml

### Antivirus False Positives:
- Several antivirus programs on VirusTotal incorrectly detect compiled Go binaries as a false positive. This issue primarily affects the Windows executable binary, but is not limited to it. If this concerns you, I recommend carefully reviewing the source code, then proceed to compile the binary yourself.
- Uploading your compiled binaries to https://virustotal.com and leaving an up-vote or a comment would be helpful as well.
