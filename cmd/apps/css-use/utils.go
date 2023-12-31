package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type Rules = *orderedmap.OrderedMap[string, string]
type Selectors = *orderedmap.OrderedMap[string, Rules]

func parseCSS(cssStr string) Selectors {
	// Initialize the CSS parser
	p := css.NewParser(parse.NewInput(bytes.NewBufferString(cssStr)), false)

	selectors := orderedmap.New[string, Rules]()

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
			printValues("BeginRulesetGrammar", valStr, selector, prop, mediaRule)
			if mediaRule != "" {
				selector = mediaRule + " " + valStr
			} else {
				selector = valStr

			}
			if _, ok := selectors.Get(selector); !ok {
				rules := orderedmap.New[string, string]()
				selectors.Set(selector, rules)
			}

		case css.EndRulesetGrammar:
			printValues("EndRulesetGrammar", valStr, selector, prop, mediaRule)
			selector = ""

		case css.BeginAtRuleGrammar:
			printValues("BeginAtRuleGrammar", valStr, selector, prop, mediaRule)
			mediaRule = prop + valStr
		case css.EndAtRuleGrammar:
			printValues("EndAtRuleGrammar", valStr, selector, prop, mediaRule)
			mediaRule = ""

		case css.AtRuleGrammar:
			printValues("AtRuleGrammar", valStr, selector, prop, mediaRule)
		case css.QualifiedRuleGrammar:
			printValues("QualifiedRuleGrammar", valStr, selector, prop, mediaRule)

		case css.DeclarationGrammar,
			css.CustomPropertyGrammar:
			printValues("DeclarationGrammar", valStr, selector, prop, mediaRule)
			if selector == "" {
				selector = mediaRule
				if selector == "@font-face" {
					continue
				}
				if _, ok := selectors.Get(selector); !ok {
					rules := orderedmap.New[string, string]()
					selectors.Set(selector, rules)
				}
			}
			if selector == "@font-face" {
				continue
			}

			rules, ok := selectors.Get(selector)
			if !ok {
				rules = orderedmap.New[string, string]()
				selectors.Set(selector, rules)
			}
			rules.Set(prop, valStr)

		case css.CommentGrammar:
			continue

		case css.ErrorGrammar, css.TokenGrammar:
			// printValues("ErrorGrammar/TokenGrammar", valStr, selector, prop, mediaRule)

		default:
			fmt.Println("Unknown grammar type:", gt)
		}
	}

	return selectors
}

func printValues(name string, valStr string, selector string, prop string, mediaRule string) {
	//_, _ = fmt.Printf("%s\n\tvalStr: %s\n\tselector: %s\n\tprop: %s\n\tmediaRule: %s\n",
	//	name, valStr, selector, prop, mediaRule)
}

func ReaderUrlOrFile(url string) (io.ReadCloser, error) {
	var reader io.ReadCloser
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		reader = resp.Body
	} else {
		file, err := os.Open(url)
		if err != nil {
			return nil, err
		}
		reader = file
	}
	return reader, nil
}

func containsAnyGlob(globHaystack []string, needles []string) bool {
	for _, needle := range needles {
		if containsGlob(globHaystack, needle) {
			return true
		}
	}
	return false
}

func containsGlob(globHaystack []string, needle string) bool {
	for _, glob := range globHaystack {
		if match, _ := path.Match(glob, needle); match {
			return true
		}
	}
	return false
}
