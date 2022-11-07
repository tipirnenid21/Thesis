// Code from "Concurrency in Go", Chapter 3
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg, wg2 sync.WaitGroup
	sayHello := func() {
		defer wg.Done()
		fmt.Println("hello")
	}
	wg.Add(1)
	wg2.Add(1)
	go sayHello()
	wg.Wait()
}
