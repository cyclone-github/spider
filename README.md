# Cyclone's URL Spider

![image](https://i.imgur.com/qc6XviF.png)

Wordlist & ngram creation tool to crawl a given url and create a wordlist / ngrams (depending on flags given).
### Usage Instructions:
- To create a simple wordlist from a specified url (will save deduplicated wordlist to url_wordlist.txt):
  - ./spider.bin -url https://github.com/cyclone-github
- To create ngrams, use flag "-ngram" and set ngram level such as "-ngram 2" or a range "-ngram 1-3"
  - ./spider.bin -url https://github.com/cyclone-github -ngram 1-3
- To set url crawl url depth of 2, use flag "-crawl 2"
  - ./spider.bin -url https://github.com/cyclone-github -crawl 2
- To set a custom output file, use flag "-o filename"
  - ./spider.bin -url https://github.com/cyclone-github -o wordlist.txt
- To set a delay to keep from being rate-limited, use flag "-delay nth" where nth is time in milliseconds
  - ./spider.bin -url https://github.com/cyclone-github -delay 100

### Compile from source code info:
- https://github.com/cyclone-github/scripts/blob/main/intro_to_go.txt
