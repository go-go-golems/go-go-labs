package cmds

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

type SummarizeLinksCommand struct {
	*cmds.CommandDescription
}

func NewSummarizeLinksCommand() (*SummarizeLinksCommand, error) {
	typeFilePath := parameters.NewParameterDefinition(
		"typeFile",
		parameters.ParameterTypeFile,
		parameters.WithHelp("File with markdown content"),
		parameters.WithRequired(true),
	)

	return &SummarizeLinksCommand{
		CommandDescription: cmds.NewCommandDescription(
			"summarize-links",
			cmds.WithShort("Fetch, convert and summarize links from markdown"),
			cmds.WithFlags(typeFilePath),
			cmds.WithArguments(),
		),
	}, nil
}

func (c *SummarizeLinksCommand) Run(ps map[string]interface{}) error {
	markdownFilePath := ps["typeFile"].(string)

	// Extract all markdown links
	links, err := extractLinks(markdownFilePath)
	if err != nil {
		return err
	}

	for _, link := range links {
		htmlContent, err := fetchHTML(link)
		if err != nil {
			return err
		}

		markdownContent := htmlToMarkdown(htmlContent)
		tempFilePath := fmt.Sprintf("temp_%d.md", time.Now().UnixNano())
		err = os.WriteFile(tempFilePath, []byte(markdownContent), 0644)
		if err != nil {
			return err
		}
		defer func(name string) {
			_ = os.Remove(name)
		}(tempFilePath)

		summary, err := summarizeContent(tempFilePath)
		if err != nil {
			return err
		}

		fmt.Println(summary)

	}

	return nil
}

func extractLinks(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Markdown link regex
	re := regexp.MustCompile(`\[(?:[^\[\]]*)\]\((https?://[^\)]+)\)`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	var links []string
	for _, match := range matches {
		links = append(links, match[1])
	}

	return links, nil
}

func fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func htmlToMarkdown(html string) string {
	// For simplicity, we're just stripping the HTML tags to convert to text
	// For a full-fledged HTML to Markdown conversion, consider using a dedicated library like 'blackfriday'
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	return doc.Text()
}

func summarizeContent(filePath string) (string, error) {
	cmd := exec.Command("pinocchio", "general", "summarize", "--article", filePath)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
