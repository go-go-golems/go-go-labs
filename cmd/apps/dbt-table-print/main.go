package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <jinja_sql_file>")
		os.Exit(1)
	}

	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	sourceRegex := regexp.MustCompile(`{{\s*source\([^,]+,\s*['"]([^']+)['"]\)\s*}}`)
	refRegex := regexp.MustCompile(`{{\s*ref\(['"]([^)]+)['"]\)\s*}}`)

	sources := make(map[string]bool)
	refs := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		for _, match := range sourceRegex.FindAllStringSubmatch(line, -1) {
			sources[strings.TrimSpace(match[1])] = true
		}
		for _, match := range refRegex.FindAllStringSubmatch(line, -1) {
			refs[strings.TrimSpace(match[1])] = true
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	fmt.Println("sources:")
	for source := range sources {
		fmt.Printf("- %s\n", source)
	}
	fmt.Println("refs:")
	for ref := range refs {
		fmt.Printf("- %s\n", ref)
	}
}
