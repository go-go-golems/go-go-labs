package main

import (
	"fmt"
	"runtime"
	"time"
)

type Data struct {
	Text string
	Blob []byte // Add a slice to hold approximately 100 kB
}

var dataMap = make(map[int]*Data)
var counter int

func testFunction() *Data {
	// Allocate approximately 100 kB for each Data instance
	data := &Data{
		Text: "This is some text",
		Blob: make([]byte, 100000), // 100,000 bytes ~ 100 kB
	}
	return data
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB\n", m.Alloc/1024/1024)
}

func main() {
	// run a goroutine that prints the memory usage every second
	go func() {
		for {
			printMemUsage()
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		go func() {
			data := testFunction()
			// Protect access to global map and counter with a mutex or similar
			// to avoid concurrent map writes and race conditions.
			// This example does not include synchronization mechanisms for simplicity.
			dataMap[counter] = data
			counter++
		}()
		time.Sleep(10 * time.Millisecond)
	}

}
