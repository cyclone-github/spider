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
