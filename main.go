package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	xurls "mvdan.cc/xurls/v2"
)

type result struct {
	Body string
	Urls []string
}

var m sync.Mutex
var wg sync.WaitGroup

type webFetcher map[string]*result

func (f *webFetcher) Fetch(url string) (body string, urls []string, err error) {
	fmt.Println("fetching: ", url)
	res, err := http.Get(url)

	if err != nil {
		err = fmt.Errorf("couldn't fetch url: %s, error: %s", url, err.Error())
		fmt.Println(err)

		return "", nil, err
	}

	bodyBytes, _ := io.ReadAll(res.Body)
	res.Body.Close()

	body = string(bodyBytes)

	rxStrict := xurls.Strict()
	urls = rxStrict.FindAllString(body, -1)

	return body, urls, nil
}

func Crawl(url string, depth int, fetcher webFetcher) {
	if depth <= 0 {
		return
	}

	body, urls, err := fetcher.Fetch(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	m.Lock()

	fetcher[url] = &result{
		Body: body,
		Urls: urls,
	}

	m.Unlock()

	for _, u := range urls {
		m.Lock()
		if _, ok := fetcher[u]; !ok {
			m.Unlock()
			wg.Add(1)

			go func(url string) {
				defer wg.Done()

				Crawl(url, depth-1, fetcher)
			}(u)
		}
	}

	return
}

func main() {
	wf := webFetcher{}

	Crawl("https://gobyexample.com/waitgroups", 1, wf)

	wg.Wait()

	PrettyPrint(wf)
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}
