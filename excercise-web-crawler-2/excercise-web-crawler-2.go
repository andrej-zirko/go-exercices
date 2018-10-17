package main

import (
	"fmt"
	"sync"
)

type MyMap struct {
	sync.RWMutex
	internalMap map[string]string
}

func (m *MyMap) add(key string) {
	defer m.Unlock()
	m.Lock()
	m.internalMap[key] = "taken"
}

func (m *MyMap) exists(key string) bool {
	defer m.RUnlock()
	m.RLock()

	_, exists := m.internalMap[key]

	return exists
}

func NewMyMap() *MyMap {
	return &MyMap{internalMap: make(map[string]string)}
}

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(ch chan string, url string, depth int, fetcher Fetcher, myMap *MyMap) {
	defer close(ch)

	if myMap.exists(url) {
		return
	}

	myMap.add(url)

	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		ch <- err.Error()
		return
	}
	ch <- fmt.Sprintf("found: %s %q\n", url, body)

	var subChannels []chan string
	for _, u := range urls {
		subChannel := make(chan string)
		subChannels = append(subChannels, subChannel)
		go Crawl(subChannel, u, depth-1, fetcher, myMap)
	}

	for _, subChannel := range subChannels {
		for foundURL := range subChannel {
			ch <- foundURL
		}
	}

	return
}

func main() {
	for i := 0; i < 10000; i++ {
		execute()
	}
}

func execute() {
	ch := make(chan string)
	myMap := NewMyMap()
	go Crawl(ch, "https://golang.org/", 4, fetcher, myMap)

	for i := range ch {
		fmt.Println(i)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
