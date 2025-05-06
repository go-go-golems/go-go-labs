// https://chatgpt.com/c/681a1825-72e0-8012-9112-5fbc9367793e
// File: main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

/* ---------- core types & constants ---------- */

const defaultMirror = "https://libgen.rs"

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

var mirror string
var debug bool

/* ---------- CLI wiring ---------- */

func main() {
	rootCmd := &cobra.Command{
		Use:   "libgen",
		Short: "Tiny LibGen helper in Go",
	}

	rootCmd.PersistentFlags().StringVar(&mirror, "mirror", defaultMirror, "LibGen mirror")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print raw URLs, JSON, and download links")

	/* --- search --- */
	var query string
	var limit int
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search books",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(query, limit)
		},
	}
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "search query (required)")
	searchCmd.Flags().IntVarP(&limit, "limit", "n", 10, "max results")
	searchCmd.MarkFlagRequired("query")
	rootCmd.AddCommand(searchCmd)

	/* --- download --- */
	var id string
	var out string
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download PDF for a given LibGen ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(id, out)
		},
	}
	downloadCmd.Flags().StringVarP(&id, "id", "i", "", "book ID (required)")
	downloadCmd.Flags().StringVarP(&out, "out", "o", "", "output filename (defaults to <md5>.pdf)")
	downloadCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(downloadCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

/* ---------- search flow ---------- */

func runSearch(q string, n int) error {
	searchURL, ids, err := crawlIDs(q, n)
	if err != nil {
		return err
	}

	jsonURL, raw, books, err := fetchDetails(ids)
	if err != nil {
		return err
	}

	if debug {
		fmt.Println("# search.php URL:")
		fmt.Println(searchURL)
		fmt.Println("# json.php URL:")
		fmt.Println(jsonURL)
		fmt.Println("# raw JSON:")
		var pretty []byte
		pretty, _ = json.MarshalIndent(raw, "", "  ")
		fmt.Println(string(pretty))
		fmt.Println("# download links:")
		for _, b := range books {
			fmt.Printf("https://books.ms/main/%s\n", b.MD5)
		}
		fmt.Println()
	}

	/* pretty (concise) output */
	for _, b := range books {
		fmt.Printf("%-7s  %s — %s (%s) [%s]\n",
			b.ID, b.Title, b.Author, b.Year, b.MD5)
	}
	return nil
}

/* ---------- download flow ---------- */

func runDownload(id, out string) error {
	_, _, books, err := fetchDetails([]string{id})
	if err != nil {
		return err
	}
	if len(books) == 0 {
		return fmt.Errorf("no book found for id %s", id)
	}
	b := books[0]
	initialURL := fmt.Sprintf("https://books.ms/main/%s", b.MD5)

	if debug {
		fmt.Println("# initial book page URL:")
		fmt.Println(initialURL)
	}

	// Fetch the initial page to find the actual download link
	resp, err := http.Get(initialURL)
	if err != nil {
		return fmt.Errorf("failed to get initial book page %s: %w", initialURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for debugging if possible
		return fmt.Errorf("failed to get initial book page %s: status %s, body: %s", initialURL, resp.Status, string(bodyBytes))
	}

	pageDoc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML from %s: %w", initialURL, err)
	}

	var foundHref string
	// Look for a link that points to download.books.ms and has a path structure like /main/PART1/PART2/FILENAME
	pageDoc.Find("a[href]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if !exists {
			return true // continue
		}

		parsedHref, err_parse := url.Parse(href)
		if err_parse != nil {
			if debug {
				fmt.Printf("# skipping malformed href: %s, error: %v\n", href, err_parse)
			}
			return true // continue, malformed href
		}

		// Check if the host is download.books.ms or if it's a relative link that contains "download.books.ms/main/"
		hostIsDownloadBooksMs := parsedHref.Host == "download.books.ms"
		isLikelyRelativeDownloadLink := parsedHref.Host == "" && strings.Contains(href, "download.books.ms/main/")

		if hostIsDownloadBooksMs || isLikelyRelativeDownloadLink {
			pathTrimmed := strings.Trim(parsedHref.Path, "/")
			pathParts := strings.Split(pathTrimmed, "/")

			mainIndex := -1
			for idx, part := range pathParts {
				if part == "main" {
					mainIndex = idx
					break
				}
			}

			// We expect a structure like /main/PART1/PART2/FILENAME.EXT
			// This means at least 3 segments after "main".
			// So, pathParts[mainIndex], pathParts[mainIndex+1], pathParts[mainIndex+2], pathParts[mainIndex+3] must exist.
			// len(pathParts) must be > mainIndex + 3
			if mainIndex != -1 && len(pathParts) > mainIndex+3 {
				foundHref = href
				if debug {
					fmt.Printf("# matched download href: %s with host '%s' and path '%s'\n", href, parsedHref.Host, parsedHref.Path)
				}
				return false // stop searching
			} else if debug && mainIndex != -1 {
				fmt.Printf("# href '%s' matched host/path prefix but not depth: pathParts len %d, mainIndex %d\n", href, len(pathParts), mainIndex)
			}
		}
		return true // continue searching
	})

	if foundHref == "" {
		return fmt.Errorf("could not find a suitable download link on page %s. Looked for links to 'download.books.ms/main/' with at least 3 path segments after 'main' (e.g., /main/part1/part2/filename.ext)", initialURL)
	}

	// Resolve the found href against the page's URL (after redirects)
	pageFinalURL := resp.Request.URL
	actualDownloadURL, err := pageFinalURL.Parse(foundHref)
	if err != nil {
		return fmt.Errorf("failed to resolve download link '%s' against page URL '%s': %w", foundHref, pageFinalURL.String(), err)
	}

	actualDownloadURLString := actualDownloadURL.String()

	if debug {
		fmt.Println("# found raw download href:", foundHref)
		fmt.Println("# page final URL for resolving relative links:", pageFinalURL.String())
		fmt.Println("# resolved actual download URL:")
		fmt.Println(actualDownloadURLString)
	}

	if out == "" {
		pathSegments := strings.Split(actualDownloadURL.Path, "/")
		if len(pathSegments) > 0 {
			filename := pathSegments[len(pathSegments)-1]
			decodedFilename, err_decode := url.PathUnescape(filename)
			if err_decode == nil && decodedFilename != "" {
				out = decodedFilename
			} else {
				out = b.MD5 + "." + b.Extension
				if debug && err_decode != nil {
					fmt.Printf("# failed to decode filename '%s' from URL path: %v. Falling back to MD5.extension\n", filename, err_decode)
				} else if debug && decodedFilename == "" {
					fmt.Printf("# filename '%s' from URL path is empty after decoding. Falling back to MD5.extension\n", filename)
				}
			}
		} else {
			out = b.MD5 + "." + b.Extension
			if debug {
				fmt.Println("# URL path has no segments to extract filename. Falling back to MD5.extension")
			}
		}
		if debug {
			fmt.Printf("# determined output filename: %s\n", out)
		}
	}
	return downloadFile(actualDownloadURLString, out)
}

/* ---------- libgen helpers ---------- */

func crawlIDs(query string, limit int) (string, []string, error) {
	v := url.Values{
		"req":    {query},
		"res":    {fmt.Sprint(limit)},
		"view":   {"simple"},
		"phrase": {"1"},
		"column": {"def"},
	}
	searchURL := fmt.Sprintf("%s/search.php?%s", mirror, v.Encode())

	doc, err := goquery.NewDocument(searchURL)
	if err != nil {
		return "", nil, err
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
	return searchURL, ids, nil
}

func fetchDetails(ids []string) (string, []map[string]string, []Book, error) {
	fields := "id,title,author,year,extension,filesize,md5,pages"
	jsonURL := fmt.Sprintf("%s/json.php?object=libgen&ids=%s&fields=%s",
		mirror, strings.Join(ids, ","), fields)

	resp, err := http.Get(jsonURL)
	if err != nil {
		return "", nil, nil, err
	}
	defer resp.Body.Close()

	var raw []map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", nil, nil, err
	}

	books := make([]Book, 0, len(raw))
	for _, m := range raw {
		books = append(books, Book{
			ID:        m["id"],
			Title:     m["title"],
			Author:    m["author"],
			Year:      m["year"],
			Pages:     m["pages"],
			Extension: m["extension"],
			MD5:       m["md5"],
			FileSize:  m["filesize"],
		})
	}
	return jsonURL, raw, books, nil
}

func downloadFile(url, out string) error {
	p := tea.NewProgram(newDownloadModel(url, out, "Starting download...", nil))

	// Update the model with the program reference
	if m, ok := p.Model().(downloadModel); ok {
		m.program = p
		p.SetModel(m)
	}

	m := p.Start()

	// Check if the final model indicates an error
	if finalModel, ok := m.(downloadModel); ok && finalModel.err != nil {
		return finalModel.err
	}
	return nil
}
