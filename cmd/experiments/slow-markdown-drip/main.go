package main

import (
	"fmt"
	"time"
)

func main() {
	markdown := `# Test Markdown

This is a test markdown file with code blocks.

## First Code Block

` + "```python" + `
def hello_world():
    print("Hello, World!")

hello_world()
` + "```" + `

Some text between code blocks.

## Second Code Block

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello from Go!")
}
` + "```" + `

End of the markdown file.`

	lines := splitIntoLines(markdown)

	for _, line := range lines {
		fmt.Println(line)
		time.Sleep(200 * time.Millisecond)
	}
}

func splitIntoLines(s string) []string {
	var lines []string
	var currentLine string

	for _, char := range s {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
