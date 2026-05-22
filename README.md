[![Readme Card](https://github-readme-stats-fast.vercel.app/api/pin/?username=cyclone-github&repo=spider&theme=gruvbox)](https://github.com/cyclone-github/spider)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/spider)](https://goreportcard.com/report/github.com/cyclone-github/spider)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/issues)
[![License](https://img.shields.io/github/license/cyclone-github/spider.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/spider.svg)](https://pkg.go.dev/github.com/cyclone-github/spider)

Spider is a web crawler and wordlist/ngram generator written in Go that crawls specified URLs or local files to produce frequency-sorted wordlists and ngrams. Users can customize crawl depth, output files, frequency sort, and ngram options, making it ideal for web scraping to create targeted wordlists for tools like hashcat or John the Ripper. Since Spider is written in Go, it requires no additional libraries to download or install.
*Spider just works*.

### Install latest release:
```
go install github.com/cyclone-github/spider@latest
```
### Install from latest source code (bleeding edge):
```
go install github.com/cyclone-github/spider@main
```

### Modes
- **URL mode** (`-url`) — crawl a website and create wordlist/ngrams (frequency sorted optional)
- **File mode** (`-file`) — process a local text file to create wordlist/ngrams (frequency sorted optional)

# Spider: URL Mode
```
spider -url 'https://github.com/hashpwn' -crawl 2 -delay 20 -sort -ngram 1-3 -timeout 10 -url-match 'hashpwn' -text-match 'hashpwn' -o hashpwn_spider.txt -agent 'foobar agent'
```
```
 ------------------ 
| Cyclone's Spider |
 ------------------ 

Crawling URL:   https://github.com/hashpwn
Base domain:    github.com
Crawl depth:    2
ngram len:      1-3
Crawl delay:    20ms (increase to avoid rate limiting)
Timeout:        10 sec
URL match:      hashpwn
Text match:     hashpwn
Scan/match:     22/21
Unique words:   847
Unique ngrams:  3816
Sorting wordlist by frequency...
Writing...      [====================] 100.00%
Output file:    hashpwn_spider.txt
RAM used:       0.003 GB
Runtime:        11.279s
```

When `-text-match` is used, all pages are still crawled for URL discovery but only pages with matching text are added to the wordlist. Crawl progress shows scanned vs matched:
```
spider -url 'https://en.wikipedia.org/wiki/PBKDF2' -crawl 2 -sort -text-match 'pbkdf2' -delay 10 -o pbkdf2_spider.txt
```
```
 ------------------ 
| Cyclone's Spider |
 ------------------ 

Crawling URL:   https://en.wikipedia.org/wiki/PBKDF2
Base domain:    en.wikipedia.org
Crawl depth:    2
ngram len:      1
Crawl delay:    10ms (increase to avoid rate limiting)
Timeout:        10 sec
Text match:     pbkdf2
Scan/match:     213/114
Unique words:   34539
Unique ngrams:  34539
Sorting wordlist by frequency...
Writing...      [====================] 100.00%
Output file:    pbkdf2_spider.txt
RAM used:       0.012 GB
Runtime:        13.715s
```

# Spider: File Mode
```
spider -file kjv_bible.txt -sort -ngram 1-3
```
```
 ------------------ 
| Cyclone's Spider |
 ------------------ 

Reading file:   kjv_bible.txt
ngram len:      1-3
Processing...   [====================] 100.00%
Unique words:   35412
Unique ngrams:  877394
Sorting wordlist by frequency...
Writing...      [====================] 100.00%
Output file:    kjv_bible_spider.txt
RAM used:       0.073 GB
Runtime:        1.137s
```

Wordlist & ngram creation tool to crawl a given url or process a local file to create wordlists and/or ngrams (depending on flags given).
### Usage Instructions:
- To create a simple wordlist from a specified url (will save deduplicated wordlist to url_spider.txt if `-o` is not set):
  - `spider -url 'https://github.com/cyclone-github'`
- To set url crawl depth of 2 and create ngrams len 1-5, use flag "-crawl 2" and "-ngram 1-5"
  - `spider -url 'https://github.com/cyclone-github' -crawl 2 -ngram 1-5`
- To set a custom output file, use flag "-o filename"
  - `spider -url 'https://github.com/cyclone-github' -o wordlist.txt`
- To set a delay to keep from being rate-limited, use flag "-delay nth" where nth is time in milliseconds
  - `spider -url 'https://github.com/cyclone-github' -delay 100`
- To set a URL timeout, use flag "-timeout nth" where nth is time in seconds (default 10)
  - `spider -url 'https://github.com/cyclone-github' -timeout 10`
- To create ngrams len 1-3 and sort output by frequency, use "-ngram 1-3" "-sort"
  - `spider -url 'https://github.com/cyclone-github' -ngram 1-3 -sort`
- To filter crawled URLs by keyword "spider" (only follow/crawl matching URLs)
  - `spider -url 'https://github.com/cyclone-github' -url-match 'spider'`
- Only match pages containing text keyword (all URLs are still crawled, but only pages containing keyword are added to wordlist)
  - `spider -url 'https://en.wikipedia.org/wiki/PBKDF2' -text-match 'pbkdf2'`
- To specify a custom user-agent
  - `spider -url 'https://github.com/cyclone-github' -agent 'foobar user agent'`
- To process a local text file, create ngrams len 1-3 and sort output by frequency
  - `spider -file foobar.txt -ngram 1-3 -sort`
- Run `spider -help` to see a list of all options

### spider -help
```
Usage of spider:
  -agent string
        Custom user-agent (default "Mozilla/5.0 (X11) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36 Spider/1.0.0")
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
  -text-match string
        Only process pages with text containing this keyword (case-insensitive); all URLs are still crawled
  -timeout int
        Timeout for URL crawling in seconds (default 10)
  -url string
        URL of the website to scrape
  -url-match string
        Only crawl URLs containing this keyword (case-insensitive)
  -version
        Display version
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
