package main

import (
	"log"
	"os"
)

func main() {
	if err := Execute(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
