// Code from "Concurrency in Go", Chapter 3
package main

import "fmt"

func main() {
	go func() {
		fmt.Println("hello")
	}()
}
