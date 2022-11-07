package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// We use the function DoRequest to do a HTTP request to certain url.
func DoRequest(url string) (int, error) {
	// Create a HTTP client with a timeout of 10 second.
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	// Do a HTTP request to the url.
	resp, err := client.Get(url)
	if err != nil {
		return 0, err // return error if any
	}
	defer resp.Body.Close()                    // close the response body after we finish reading it
	log.Printf("%s: %d", url, resp.StatusCode) // print the status code
	return resp.StatusCode, nil
}

// 1. run one by one, which is very slow
func SequentialRun(URLS []string) {
	for _, url := range URLS {
		code, _ := DoRequest(url)       // do the request for every url
		log.Printf("%s: %d", url, code) // print the result
	}
}

// 2. go faster using unbounded goroutine and synchronized wait group
func GoroutineRun(URLS []string) {
	var wg sync.WaitGroup // instantiate a wait group
	for _, url := range URLS {
		wg.Add(1) // add one to the wait group
		go func(url string) {
			defer wg.Done() // when the goroutine is done, decrease the wait group by one
			DoRequest(url)
		}(url)
	}
	wg.Wait() // wait until the wait group is decreased to 0
}

// 3. workerpool pattern
// in this part, we create a worker pool with 10 workers and do the request concurrently.
type HTTPResponse struct { // define a struct to store the result
	code int    // status code
	url  string // url
}

// worker function
func worker(urlChan <-chan string, resultsChan chan<- HTTPResponse) {
	for url := range urlChan { // read the url from the urlChan channel
		code, _ := DoRequest(url)              // do the request for every url
		resultsChan <- HTTPResponse{code, url} // write the result into the resultsChan channel
	}
}
func WorkerPool(URLS []string) {
	var numWorkers int = 10 // number of workers

	urlChan := make(chan string, numWorkers) // channel to store the url
	resultsChan := make(chan HTTPResponse)   // channel to store the result
	for i := 0; i < numWorkers; i++ {        // create numWorkers workers
		go worker(urlChan, resultsChan) // pass the urlChan and resultsChan to the worker function
	}

	// write url into the urlChan
	go func() {
		for _, url := range URLS { // read the url line by line
			urlChan <- url // write the url into the urlChan channel
		}
	}()

	for res := range resultsChan { // read the result from the resultsChan channel
		log.Printf("%s: %d", res.url, res.code) // print the result
	}

	close(urlChan)     // close the urlChan channel
	close(resultsChan) // close the resultsChan channel
}

// use mutex store the result into a map
func MutexRun(URLS []string) {
	result := make(map[string]int) // create a map to store the result
	var mutex sync.Mutex           // instantiate a mutex

	var wg sync.WaitGroup      // instantiate a wait group
	for _, url := range URLS { // read the url line by line
		wg.Add(1) // add one to the wait group
		go func(url string) {
			defer wg.Done()           // when the goroutine is done, decrease the wait group by one
			code, _ := DoRequest(url) // do the request for every url
			mutex.Lock()              // lock the mutex
			result[url] = code        // critical section, write the result into the map
			mutex.Unlock()            // unlock the mutex
		}(url)
	}
	wg.Wait() // wait until the wait group is decreased to 0

	fmt.Println(result) // print the result

}

func main() {
	URLS := []string{
		"http://www.youtube.com",
		"http://www.facebook.com",
		"http://www.baidu.com",
		"http://www.yahoo.com",
		"http://www.amazon.com",
		"http://www.wikipedia.org",
		"http://www.qq.com",
		"http://www.google.co.in",
		"http://www.twitter.com",
		"http://www.live.com",
		"http://www.taobao.com",
		"http://www.bing.com",
		"http://www.instagram.com",
		"http://www.weibo.com",
	}

	// four different run
	SequentialRun(URLS)
	fmt.Println("Run with goroutine:")
	GoroutineRun(URLS)
	fmt.Println("Run with workerpool:")
	WorkerPool(URLS)
	fmt.Println("Run with mutex:")
	MutexRun(URLS)

}
