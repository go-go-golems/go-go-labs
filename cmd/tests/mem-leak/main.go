package main

import (
	"fmt"
	"time"
)

type Data struct {
	Text string
}

var dataMap = make(map[int]*Data)
var counter int

func testFunction() *Data {
	data := &Data{Text: "This is some text"}
	return data
}

func main() {
	for {
		go func() {
			data := testFunction()
			// Protect access to global map and counter with a mutex or similar
			// to avoid concurrent map writes and race conditions.
			// This example does not include synchronization mechanisms for simplicity.
			dataMap[counter] = data
			counter++
			fmt.Println(data)
		}()
		time.Sleep(10 * time.Millisecond)
	}
}
