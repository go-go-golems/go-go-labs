package main

import (
	"fmt"
	"time"
)

type Data struct {
	Text string
}

func testFunction() *Data {
	data := &Data{Text: "This is some text"}
	return data
}

func main() {
	for {
		go func() {
			data := testFunction()
			fmt.Println(data)
		}()
		time.Sleep(10 * time.Millisecond)
	}
}
