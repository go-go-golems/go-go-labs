package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Docstring represents a structured format of docstrings
type Docstring struct {
	Annotations      map[string][]string
	MethodOrField    string
	FullDocstring    string
	ContentDocstring string
}

type Scanner struct {
	StartRegex     *regexp.Regexp
	EndRegex       *regexp.Regexp
	AttributeRegex *regexp.Regexp
	StripRegex     *regexp.Regexp // New field
}

func NewScanner(start, end, attribute, strip string) *Scanner {
	return &Scanner{
		StartRegex:     regexp.MustCompile(start),
		EndRegex:       regexp.MustCompile(end),
		AttributeRegex: regexp.MustCompile(attribute),
		StripRegex:     regexp.MustCompile(strip), // Initializing the new regex
	}
}

func (s *Scanner) ParseAnnotations(input string) Docstring {
	lines := strings.Split(input, "\n")
	doc := Docstring{
		Annotations:   make(map[string][]string),
		FullDocstring: input,
	}

	var contentLines []string

	for _, line := range lines {
		matches := s.AttributeRegex.FindStringSubmatch(line)
		if len(matches) > 2 {
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			doc.Annotations[key] = append(doc.Annotations[key], value)
		}

		// Using the StripRegex to clean the line
		if s.StripRegex != nil {
			cleanedLine := s.StripRegex.ReplaceAllString(line, "")
			contentLines = append(contentLines, cleanedLine)
		} else {
			contentLines = append(contentLines, line)
		}
	}

	// Remove any leading or trailing whitespace lines
	content := strings.Join(contentLines, "\n")
	content = strings.Trim(content, "\n\t ")

	doc.ContentDocstring = content
	return doc
}

func (s *Scanner) ScanFile(filePath string) ([]*Docstring, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	ret := []*Docstring{}

	inDocstring := false
	currentDocstring := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if s.StartRegex.MatchString(line) {
			inDocstring = true
			currentDocstring = ""
		}
		if inDocstring {
			currentDocstring += line + "\n"
		}
		if s.EndRegex.MatchString(line) {
			inDocstring = false
			for scanner.Scan() {
				line = scanner.Text()
				if strings.TrimSpace(line) != "" {
					doc := s.ParseAnnotations(currentDocstring)
					doc.MethodOrField = strings.TrimSpace(line)
					ret = append(ret, &doc)
					break
				}
			}
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return ret, nil
}
