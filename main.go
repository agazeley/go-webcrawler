package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Opts struct {
	maxCrawls uint64
	rootUrl   string
}

type Crawler struct {
	crawledPages    map[string]bool
	maxCrawls       uint64
	maxRoutines     int
	workChan        chan []string
	unseenLinksChan chan string
	rootUrl         string
	wg              sync.WaitGroup
}

func NewCrawler(opts *Opts) *Crawler {
	c := &Crawler{
		crawledPages:    make(map[string]bool),
		maxCrawls:       uint64(math.Inf(1)),
		workChan:        make(chan []string),
		unseenLinksChan: make(chan string),
		rootUrl:         opts.rootUrl,
		maxRoutines:     20,
		wg:              sync.WaitGroup{},
	}

	if opts.maxCrawls != 0 {
		c.maxCrawls = opts.maxCrawls
	}

	return c
}

// isValidUrl tests a string to determine if it is a well-structured url or not.
// https://golangcode.com/how-to-check-if-a-string-is-a-url/
func isValidUrl(toTest string) bool {
	if toTest == "" {
		return false
	}

	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

// Parse out all anchors
func parseResponse(resp *http.Response) []string {
	defer resp.Body.Close()
	urls := make([]string, 0)

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Error loading HTTP response body. ", err)
		return urls
	}

	document.Find("a").Each(func(i int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists && isValidUrl(href) {
			urls = append(urls, href)
		}
	})
	return urls
}

// Fetch all the anchor links on a page and push them to the channel
func scrapeUrl(url string) []string {

	resp, err := http.Get(url)
	if err != nil {
		return []string{}
	}

	urls := parseResponse(resp)
	fmt.Println(url)
	for _, link := range urls {
		fmt.Println("\t", link)
	}
	return urls
}

func (c *Crawler) checkUrls() {
	// Loop through and make sure you have not seen the URL
	n := 0
	for urls := range c.workChan {
		// Puhs unseen into the channel to process unseen links
		for _, url := range urls {

			// Process only uncrawled pages
			if !c.crawledPages[url] {
				n += 1
				c.wg.Add(1)
				c.crawledPages[url] = true
				c.unseenLinksChan <- url
			}

			// Exit condition
			if n >= int(c.maxCrawls) {
				c.wg.Wait()
				return // done
			}
		}
	}
}

func (c *Crawler) createWorkers() {
	// Set up our routines  to scape the urls
	for i := 0; i < c.maxRoutines; i++ {
		go func() {
			for url := range c.unseenLinksChan {
				foundLinks := scrapeUrl(url)

				// Push to url checking channel
				go func() { c.workChan <- foundLinks }()
				c.wg.Done()
			}
		}()
	}
}

func (c *Crawler) Crawl() map[string]bool {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for range sig {
			// sig is a ^C, handle it
			fmt.Println("Crawled", len(c.crawledPages), "pages")
			os.Exit(1)
		}
	}()

	// Submit first job
	go func() {
		c.workChan <- []string{c.rootUrl}
	}()

	c.createWorkers()
	c.checkUrls()

	return c.crawledPages
}

func parseArgs(args []string) (*Opts, error) {
	opts := &Opts{}
	if len(args) < 2 {
		return nil, fmt.Errorf("Not enough arguements supplied")
	}

	opts.rootUrl = args[1]
	if !isValidUrl(opts.rootUrl) {
		return nil, fmt.Errorf("Invalid URL format")
	}

	if len(args) == 3 {
		opts.maxCrawls, _ = strconv.ParseUint(args[2], 0, 64)
	}

	return opts, nil
}

func main() {
	opts, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	crawler := NewCrawler(opts)
	http.DefaultClient.Timeout = time.Second * 10
	crawledPages := crawler.Crawl()

	fmt.Println("Crawled", len(crawledPages), "pages")
	os.Exit(0)
}
