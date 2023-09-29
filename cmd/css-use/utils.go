package main

import (
	"bytes"
	"fmt"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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
