package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <html file path>")
		return
	}

	filePath := os.Args[1]
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	doc, err := html.Parse(strings.NewReader(string(fileContent)))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	var cssContents []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "style" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					cssContents = append(cssContents, c.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	for _, css := range cssContents {
		fmt.Println("Inline CSS:")
		fmt.Println("-----------")
		v := parseCSS(css)
		v_, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Println("Error marshalling:", err)
			return
		}
		_ = v
		fmt.Println(string(v_))
		//fmt.Println(css)
		fmt.Println("-----------")
	}
}

func parseCSS(cssStr string) map[string]map[string]string {
	// Initialize the CSS parser
	p := css.NewParser(parse.NewInput(bytes.NewBufferString(cssStr)), false)

	rules := make(map[string]map[string]string)

	selector := ""
	mediaRule := ""

	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			break
		}
		values := p.Values()

		prop := string(data)

		valStr := ""
		for _, val := range values {
			valStr += string(val.Data)
		}

		switch gt {
		case css.BeginRulesetGrammar:
			if mediaRule != "" {
				selector = mediaRule + " " + selector
			} else {
				selector = valStr

			}
			rules[selector] = make(map[string]string)
		case css.EndRulesetGrammar:

		case css.BeginAtRuleGrammar:
			mediaRule = prop + valStr
		case css.EndAtRuleGrammar:
			fmt.Println("EndAtRuleGrammar:", valStr)
			mediaRule = ""

		case css.AtRuleGrammar,
			css.QualifiedRuleGrammar:
			fmt.Println("QualifiedRuleGrammar:", valStr)

		case css.DeclarationGrammar,
			css.CustomPropertyGrammar:
			if selector == "" {
				selector = mediaRule
				if selector == "@font-face" {
					continue
				}
				rules[selector] = make(map[string]string)
			}
			if selector == "@font-face" {
				continue
			}
			rules[selector][prop] = valStr

		case css.CommentGrammar:
			continue

		default:
			fmt.Println("Unknown grammar type:", gt)
		}
	}

	return rules
}
