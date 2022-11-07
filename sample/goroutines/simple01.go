// Code from "Concurrency in Go", Chapter 3
package main

import "fmt"

func main() {
	go sayHello()
	// continue doing other things
}

func sayHello() {
	fmt.Println("hello")
}
