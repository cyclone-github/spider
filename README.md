# Cyclone's URL Spider
Website spider to crawl website and create a wordlist that is sorted by frequency

Usage Instructions:
- ./spider.bin -url https://github.com/cyclone-github
- to set crawl depth, use flag "-crawl" (defaults to -crawl 1)
- to set phrase depth, use flag "-phrase" (defaults to -phrase 1)
- ./spider -url https://github.com/cyclone-github -crawl 2 -phrase 2
- defaults output to url_wordlist.txt, but can also be specified by flag "-o"

Compile from source code info:
- https://github.com/cyclone-github/scripts/blob/main/intro_to_go.txt
