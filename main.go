package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

const (
	basePage   = "https://www.ptt.cc/bbs/Beauty/"
	homePage   = "https://www.ptt.cc/bbs/Beauty/index2061.html"
	workNumber = 3
    crawlerNumber = 1
)

var (
	pageFilter  = regexp.MustCompile(`(M.([^\"]+).html)"`)
	imageFilter = regexp.MustCompile(`href="(([^\"]+).jpg)"`)
)

func worker(linkChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for link := range linkChan {
		log.Println(link)
		res, err := http.Get(link)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()
		fp, err := os.Create(link[len(link)-8:])
		if err != nil {
			log.Println(err)
		}
		if _, err := io.Copy(fp, res.Body); err != nil {
			log.Println(err)
		}

	}
}

func imageURLParser(htmlBody string, linkChan chan<- string) {
	imageURLs := imageFilter.FindAllStringSubmatch(htmlBody, -1)
	for _, imageURL := range imageURLs {
        linkChan <- string(imageURL[1][:])
	}
}

func getHtmlBody(URL string) (string, error) {
	res, err := http.Get(URL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body_bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body_bytes), nil
}

func crawler(pageLinkChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	linkChan := make(chan string)
	_wg := &sync.WaitGroup{}
	for i := 0; i < workNumber; i++ {
		_wg.Add(1)
		go worker(linkChan, _wg)
	}
	for pageURL := range pageLinkChan {
		body, err := getHtmlBody(pageURL)
		if err != nil {
			log.Fatal(err)
		}
		imageURLParser(body, linkChan)
	}
	close(linkChan)
	_wg.Wait()
}

func pageURLParser(htmlURL string) {
	wg := &sync.WaitGroup{}
	linkChan := make(chan string)
	for i := 0; i < crawlerNumber; i++ {
		wg.Add(1)
		go crawler(linkChan, wg)
	}
	body, err := getHtmlBody(htmlURL)
	if err != nil {
		log.Fatal(err)
	}
	pageURLs := pageFilter.FindAllStringSubmatch(body, -1)
	for _, pageURL := range pageURLs {
		pageURL_copy := string(pageURL[1][:])
		log.Println(basePage + pageURL_copy)
		linkChan <- (basePage + pageURL_copy)
	}
	close(linkChan)
	wg.Wait()
}

func main() {
	pageURLParser(homePage)
}
