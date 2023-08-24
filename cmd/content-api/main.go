package main

import (
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve"
	"strings"
)

func main() {
	// open a new index
	mapping := bleve.NewIndexMapping()

	index, err := bleve.Open("example.bleve")
	if err == bleve.ErrorIndexPathDoesNotExist {
		index, err = bleve.New("example.bleve", mapping)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if err != nil {
		fmt.Println(err)
		return
	}

	// output the index mapping as json to stdout
	buf := strings.Builder{}
	// pretty print json
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(mapping)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(buf.String())

	data := struct {
		Name string
	}{
		Name: "text",
	}

	// index some data
	err = index.Index("id", data)
	if err != nil {
		panic(err)
	}

	// search for some text
	query := bleve.NewMatchQuery("text")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(searchResults)
}
