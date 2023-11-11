package main

import (
	"fmt"
	"os"

	markdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var divClass string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "htmlparser [html file]",
		Short: "HTML Parser extracts content from specified divs",
		Args:  cobra.MinimumNArgs(1),
		Run:   run,
	}

	rootCmd.PersistentFlags().StringVarP(&divClass, "div", "d", "ReactMarkdown", "Class of the div to extract")

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func run(cmd *cobra.Command, args []string) {
	filename := args[0]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	converter := markdown.NewConverter("", true, nil)

	//doc.Find("div." + divClass).Each(func(i int, s *goquery.Selection) {
	doc.Find("div." + divClass).Each(func(i int, s *goquery.Selection) {
		s.Find("div.contents").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()
			if err != nil {
				fmt.Println("Error getting HTML:", err)
				return
			}

			m_, err := converter.ConvertString(html)
			if err != nil {
				fmt.Println("Error converting HTML to Markdown:", err)
				return
			}

			if m_ == "Copy" {
				return
			}

			fmt.Println(m_ + "\n\n---\n\n")
		})
	})
}
