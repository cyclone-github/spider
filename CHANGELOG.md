### v1.0.0
```
added flag "-text-match" to filter page text matches
memory and performance optimizations for -file and -url modes
-file mode streams wordlists from disk instead of loading entire files into RAM
reduced RAM usage for large -sort wordlists
default -timeout increased from 1 to 10 seconds
progress bars, stats, and errors now write to stderr
sanitize url fragments for dedup and extension checks
updated default User-Agent
```
### v0.9.1
```
added flag "-agent" to allow user to specify custom user-agent; https://github.com/cyclone-github/spider/issues/8
```
### v0.9.0
```
added flag "-url-match" to only crawl URLs containing a specified keyword; https://github.com/cyclone-github/spider/issues/6
added notice to user if no URLs are crawled when using "-crawl 1 -url-match"
exit early if zero URLs were crawled (no processing or file output)
use custom User-Agent "Spider/0.9.0 (+https://github.com/cyclone-github/spider)"
removed clearScreen function and its imports
fixed crawl-depth calculation logic
fixed restrict link collection to .html, .htm, .txt and extension-less paths
upgraded dependencies and bumped Go version to v1.24.3
```
### v0.8.1
```
updated default -delay to 10ms
```
### v0.8.0
```
added flag "-file" to allow creating ngrams from a local plaintext file (ex: foobar.txt)
added flag "-timeout" for -url mode
added flag "-sort" which sorts output by frequency
fixed several small bugs
```
### v0.7.1
```
added progress bars to word / ngrams processing & file writing operations
added RAM usage monitoring
optimized order of operations for faster processing with less RAM
TO-DO: refactor code
```
### v0.7.0
```
added feature to allow crawling specific file extensions (html, htm, txt)
added check to keep crawler from crawling offsite URLs
added flag "-delay" to avoid rate limiting (-delay 100 == 100ms delay between URL requests)
added write buffer for better performance on large files
increased crawl depth from 5 to 100 (not recommended, but enabled for edge cases)
fixed out of bounds slice bug when crawling URLs with NIL characters
fixed bug when attempting to crawl deeper than available URLs to crawl
fixed crawl depth calculation
optimized code which runs 2.8x faster vs v0.6.x during bench testing
```
### v0.6.2
```
fixed scraping logic & ngram creations bugs
switched from gocolly to goquery for web scraping
remove dups from word / ngrams output
```
### v0.5.10
```
initial github release
```
