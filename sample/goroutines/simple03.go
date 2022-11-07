// Code from "Concurrency in Go", Chapter 3
package main

import "fmt"

func main() {
	sayHello := func() {
		fmt.Println("hello")
	}
	go sayHello()
}
