[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=cyclone-github&repo=spider&theme=gruvbox)](https://github.com/cyclone-github/spider/)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/spider)](https://goreportcard.com/report/github.com/cyclone-github/spider)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/issues)
[![License](https://img.shields.io/github/license/cyclone-github/spider.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/spider.svg)](https://pkg.go.dev/github.com/cyclone-github/spider)

# Spider: URL Mode
```
spider -url 'https://forum.hashpwn.net' -crawl 3 -delay 20 -sort -ngram 1-3 -o forum.hashpwn.net_spider.txt
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
URLs crawled:   56
Processing...   [====================] 100.00%
Unique words:   3164
Unique ngrams:  17313
Sorting n-grams by frequency...
Writing...      [====================] 100.00%
Output file:    forum.hashpwn.net_spider.txt
RAM used:       0.03 GB
Runtime:        8.634s
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
- To create ngrams len 1-3 and sort output by frequency, use "-ngram 1-3" "-sort"
  - `spider -url 'https://github.com/cyclone-github' -ngram 1-3 -sort`
- To process a local text file, create ngrams len 1-3 and sort output by frequency
  - `spider -file foobar.txt -ngram 1-3 -sort`
- Run `spider -help` to see a list of all options

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
