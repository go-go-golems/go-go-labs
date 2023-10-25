package main

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/sync/errgroup"
	"os"
)

func main() {
	// open first argument as file
	fileName := os.Args[1]
	content, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	c := make(chan []byte)
	astStream := make(chan ast.Node)

	eg := errgroup.Group{}
	eg.Go(func() error {
		// read file in chunks of 50 bytes
		defer close(c)
		for i := 0; i < len(content); i += 50 {
			if i+50 > len(content) {
				c <- content[i:]
				break
			}
			c <- content[i : i+50]
		}
		return nil
	})

	eg.Go(func() error {
		debug := false
		streamContent := []byte{}
		defer close(astStream)

		var lastNode ast.Node
		var lastStart int
		var lastStop int
		chunkCount := 0

		for chunk := range c {
			chunkCount++

			streamContent = append(streamContent, chunk...)
			document := goldmark.DefaultParser().Parse(
				text.NewReader(streamContent),
			)

			if debug {
				fmt.Printf("chunk: %d, len: %d, lastStart: %d, lastStop: %d\n", chunkCount, len(streamContent), lastStart, lastStop)
			}

			err := ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					nodeType := ""
					var segments *text.Segments

					switch v := n.(type) {
					case *ast.Heading:
						nodeType = "Heading"
						segments = v.Lines()
					case *ast.FencedCodeBlock:
						nodeType = "Fenced Code block"
						segments = v.Lines()
					case *ast.CodeBlock:
						nodeType = "Code block"
						segments = v.Lines()
					case *ast.Paragraph:
						nodeType = "Paragraph"
						segments = v.Lines()
					case *ast.List:
						nodeType = "List"
						segments = v.Lines()
					case *ast.ListItem:
						nodeType = "List item"
						cur := v.FirstChild()
						segments = &text.Segments{}
						for cur != nil {
							lines_ := cur.Lines()
							for i := 0; i < lines_.Len(); i++ {
								segments.Append(lines_.At(i))
							}
							cur = cur.NextSibling()
						}
					case *ast.Blockquote:
						nodeType = "Blockquote"
						segments = v.Lines()
					default:
						return ast.WalkContinue, nil
					}
					_ = nodeType

					if segments == nil {
						return ast.WalkContinue, nil
					}
					if segments.Len() == 0 {
						return ast.WalkContinue, nil
					}

					firstSegment := segments.At(0)
					lastSegment := segments.At(segments.Len() - 1)

					end := firstSegment.Start + 40
					if len(streamContent) < end {
						end = len(streamContent)
					}
					if end > lastSegment.Stop {
						end = lastSegment.Stop
					}
					first10Characters := string(streamContent[firstSegment.Start:end])
					if debug {
						fmt.Printf("node: %s, start: %d, stop: %d, %s\n",
							nodeType, firstSegment.Start, lastSegment.Stop, first10Characters)
						fmt.Printf("lastStart: %d, lastStop: %d\n", lastStart, lastStop)
					}

					if lastNode == nil {
						lastNode = n
						lastStart = firstSegment.Start
						lastStop = lastSegment.Stop
						return ast.WalkContinue, nil
					}

					if lastSegment.Stop <= lastStop {
						return ast.WalkSkipChildren, nil
					}

					if firstSegment.Start == lastStart {
						lastNode = n
						lastStart = firstSegment.Start
						lastStop = lastSegment.Stop
						return ast.WalkContinue, nil
					}

					//fmt.Printf("lastStop: %d, lastSegment.Start: %d\n", lastStop, lastSegment.Start)
					if firstSegment.Start >= lastStop {
						astStream <- lastNode
						lastNode = n
						lastStart = firstSegment.Start
						lastStop = lastSegment.Stop
						return ast.WalkContinue, nil
					}
				}
				return ast.WalkContinue, nil
			})

			if err != nil {
				return err
			}

			if debug {
				fmt.Println()
			}
		}

		if lastNode != nil {
			astStream <- lastNode
		}
		return nil

	})

	eg.Go(func() error {
		for node := range astStream {
			switch v := node.(type) {
			case *ast.Heading:
				fmt.Printf("Heading: %s\n", string(v.Text(content)))
			case *ast.CodeBlock:
				fmt.Printf("Code block: %s\n", string(v.Text(content)))
			case *ast.FencedCodeBlock:
				lines := v.Lines()
				content_ := ""
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					content_ += string(line.Value(content))
				}

				fmt.Printf("Fenced Code block: %s (info: %s)\n", content_, v.Info.Text(content))
			case *ast.Paragraph:
				fmt.Printf("Paragraph: %s\n", string(v.Text(content)))
			case *ast.List:
				fmt.Printf("List: %s\n", string(v.Text(content)))
			case *ast.ListItem:
				cur := v.FirstChild()
				content_ := ""
				for cur != nil {
					// if cur is ast.List, skip
					if _, ok := cur.(*ast.List); ok {
						cur = cur.NextSibling()
						continue
					}
					content_ += string(cur.Text(content))
					cur = cur.NextSibling()
				}
				fmt.Printf("List item: %s (offset: %d)\n", content_, v.Offset)
			case *ast.Blockquote:
				fmt.Printf("Blockquote: %s\n", string(v.Text(content)))
			}
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		panic(err)
	}

}
