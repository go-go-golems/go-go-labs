// https://chatgpt.com/c/681a1825-72e0-8012-9112-5fbc9367793e
// File: libgen.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const mirror = "https://libgen.rs" // swap at runtime if 503

type Book struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Year      string `json:"year"`
	Pages     string `json:"pages"`
	Extension string `json:"extension"`
	MD5       string `json:"md5"`
	FileSize  string `json:"filesize"`
}

func main() {
	books, _ := Search("Gödel Escher Bach", 10)
	for _, b := range books {
		fmt.Printf("%s  %s — %s (%s) [%s]\n",
			b.ID, b.Title, b.Author, b.Year, b.MD5)
	}
}

// Search performs full-text search, grabs the first n IDs,
// then calls json.php to enrich the records.
func Search(q string, n int) ([]Book, error) {
	ids, err := crawlIDs(q, n)
	if err != nil {
		return nil, err
	}
	return fetchDetails(ids)
}

// --- 1. scrape search.php ----------------------------------------------------
func crawlIDs(query string, limit int) ([]string, error) {
	v := url.Values{
		"req":     {query},
		"res":     {fmt.Sprint(limit)},
		"view":    {"simple"},
		"phrase":  {"1"},
		"column":  {"def"},
	}
	u := fmt.Sprintf("%s/search.php?%s", mirror, v.Encode())

	doc, err := goquery.NewDocument(u)
	if err != nil {
		return nil, err
	}
	var ids []string
	doc.Find("table.c tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 || len(ids) >= limit {
			return
		}
		if id := strings.TrimSpace(s.Find("td").First().Text()); id != "" {
			ids = append(ids, id)
		}
	})
	return ids, nil
}

// --- 2. hit json.php ---------------------------------------------------------
func fetchDetails(ids []string) ([]Book, error) {
	fields := "id,title,author,year,extension,filesize,md5,pages"
	u := fmt.Sprintf("%s/json.php?object=libgen&ids=%s&fields=%s",
		mirror, strings.Join(ids, ","), fields)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw []map[string]string // json.php always returns an array
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	books := make([]Book, 0, len(raw))
	for _, m := range raw {
		book := Book{
			ID:        m["id"],
			Title:     m["title"],
			Author:    m["author"],
			Year:      m["year"],
			Pages:     m["pages"],
			Extension: m["extension"],
			MD5:       m["md5"],
			FileSize:  m["filesize"],
		}
		books = append(books, book)
	}
	return books, nil
}

