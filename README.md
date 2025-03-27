[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=cyclone-github&repo=spider&theme=gruvbox)](https://github.com/cyclone-github/spider/)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/spider)](https://goreportcard.com/report/github.com/cyclone-github/spider)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/issues)
[![License](https://img.shields.io/github/license/cyclone-github/spider.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/spider.svg)](https://github.com/cyclone-github/spider/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/spider.svg)](https://pkg.go.dev/github.com/cyclone-github/spider)

# Cyclone's URL Spider
<!-- ![image](https://i.imgur.com/Z6RjlUv.png) -->
```
 ---------------------- 
| Cyclone's URL Spider |
 ---------------------- 

Crawling URL:   https://forum.hashpwn.net
Base domain:    forum.hashpwn.net
Crawl depth:    2
ngram len:      1-3
Crawl delay:    0ms (increase this to avoid rate limiting, ex: -delay 100)
URLs crawled:   51
Processing...   [====================] 100.00%
Unique words:   1983
Unique ngrams:  11030
Writing...      [====================] 100.00%
Output file:    forum.hashpwn.net_wordlist.txt
RAM used:       0.03 GB
Runtime:        4.949s
```

Wordlist & ngram creation tool to crawl a given url and create wordlists and/or ngrams (depending on flags given).
### Usage Instructions:
- To create a simple wordlist from a specified url (will save deduplicated wordlist to url_wordlist.txt):
  - `./spider.bin -url https://github.com/cyclone-github`
- To set url crawl url depth of 2 and create ngrams len 1-5, use flag "-crawl 2" and "-ngram 1-5"
  - `./spider.bin -url https://github.com/cyclone-github -crawl 2 -ngram 1-5`
- To set a custom output file, use flag "-o filename"
  - `./spider.bin -url https://github.com/cyclone-github -o wordlist.txt`
- To set a delay to keep from being rate-limited, use flag "-delay nth" where nth is time in milliseconds
  - `./spider.bin -url https://github.com/cyclone-github -delay 100`
- Run `./spider.bin -help` to see a list of all options

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
### Change Log:
- https://github.com/cyclone-github/spider/blob/main/CHANGELOG.md
### Mentions:
- Go Package Documentation: https://pkg.go.dev/github.com/cyclone-github/spider
- Softpedia: https://www.softpedia.com/get/Internet/Other-Internet-Related/Cyclone-s-URL-Spider.shtml

### Antivirus False Positives:
- Several antivirus programs on VirusTotal incorrectly detect compiled Go binaries as a false positive. This issue primarily affects the Windows executable binary, but is not limited to it. If this concerns you, I recommend carefully reviewing the source code, then proceed to compile the binary yourself.
- Uploading your compiled binaries to https://virustotal.com and leaving an up-vote or a comment would be helpful as well.
