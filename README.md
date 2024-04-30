[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=cyclone-github&repo=spider&theme=gruvbox)](https://github.com/cyclone-github/)
# Cyclone's URL Spider

![image](https://i.imgur.com/Z6RjlUv.png)

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
  - `git clone https://github.com/cyclone-github/spider.git`
  - `cd spider`
  - `go mod init spider`
  - `go mod tidy`
  - `go build -ldflags="-s -w" .`
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
