package main

import (
	"fmt"
	"os"

	"github.com/wmentor/epub"
)

func main() {

	fmt.Println("Reading chapter")
	err := epub.Reader("/tmp/test.epub", func(chapter string, chapterHTML []byte) bool {
		fmt.Println("chapter")
		fmt.Println(chapter)
		return true
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Print to stdout")
	// print epub to Stdout as text
	err = epub.ToTxt("/tmp/test.epub", os.Stdout)
	if err != nil {
		panic(err)
	}
}
