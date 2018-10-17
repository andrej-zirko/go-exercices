package main

import (
	"fmt"
	"sync"
)

type MyMap struct {
	sync.RWMutex
	internalMap map[string]bool
}

func (m *MyMap) add(key string, found bool) {
	defer m.Unlock()
	m.Lock()
	m.internalMap[key] = found
}

func (m *MyMap) exists(key string) bool {
	defer m.RUnlock()
	m.RLock()
	_, exists := m.internalMap[key]

	return exists
}

func NewMyMap() *MyMap {
	return &MyMap{internalMap: make(map[string]bool)}
}

func (m *MyMap) getMapCopy() map[string]bool {
	return m.internalMap
}

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, wg *sync.WaitGroup, myMap *MyMap) {
	defer wg.Done()

	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	_, urls, err := fetcher.Fetch(url)
	if err != nil {
		myMap.add(url, false)
		return
	}

	myMap.add(url, true)
	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, wg, myMap)
	}

	return
}

func execute() {
	var wg sync.WaitGroup
	myMap := NewMyMap()
	wg.Add(1)
	go Crawl("https://golang.org/", 4, fetcher, &wg, myMap)
	wg.Wait()

	for key, value := range myMap.getMapCopy() {
		fmt.Printf("%s; Found: %t \n", key, value)
	}
}

func main() {
	// Parallelism test
	for i := 0; i < 10000; i++ {
		execute()
		fmt.Println("----------------------")
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
