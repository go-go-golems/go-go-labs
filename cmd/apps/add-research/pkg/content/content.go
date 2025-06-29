package content

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Message      string
	UseClipboard bool
	AttachFiles  []string
	AskForLinks  bool
	Links        []string
}

func GetContentFromUser(config Config) (string, error) {
	var content strings.Builder
	
	// If message provided via command line, use it
	if config.Message != "" {
		content.WriteString(config.Message)
		content.WriteString("\n")
	}
	
	// If clipboard flag is set, read from clipboard
	if config.UseClipboard {
		clipContent, err := clipboard.ReadAll()
		if err != nil {
			return "", errors.Wrap(err, "failed to read from clipboard")
		}
		log.Debug().Int("length", len(clipContent)).Msg("Reading content from clipboard")
		if clipContent != "" {
			if content.Len() > 0 {
				content.WriteString("\n")
			}
			content.WriteString(clipContent)
			content.WriteString("\n")
		}
	}
	
	// Check if stdin has data available (piped input)
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		// stdin has piped data
		log.Debug().Msg("Reading content from stdin")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			content.WriteString(scanner.Text())
			content.WriteString("\n")
		}
		if err := scanner.Err(); err != nil {
			return "", errors.Wrap(err, "failed to read stdin")
		}
	} else if config.Message == "" && !config.UseClipboard {
		// No piped data, no message flag, and no clipboard flag - prompt user interactively
		fmt.Println("Enter your markdown content (press Ctrl+D when finished):")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			content.WriteString(scanner.Text())
			content.WriteString("\n")
		}
		if err := scanner.Err(); err != nil {
			return "", errors.Wrap(err, "failed to read input")
		}
	}
	
	// Attach files if specified
	if len(config.AttachFiles) > 0 {
		fileContent, err := processAttachedFiles(config.AttachFiles)
		if err != nil {
			return "", errors.Wrap(err, "failed to process attached files")
		}
		if fileContent != "" {
			if content.Len() > 0 {
				content.WriteString("\n---\n\n")
			}
			content.WriteString(fileContent)
		}
	}
	
	// Handle links
	links := config.Links
	if config.AskForLinks {
		userLinks, err := askForLinks()
		if err != nil {
			return "", errors.Wrap(err, "failed to get links from user")
		}
		links = append(links, userLinks...)
	}
	
	if len(links) > 0 {
		linkContent := processLinks(links)
		if linkContent != "" {
			if content.Len() > 0 {
				content.WriteString("\n---\n\n")
			}
			content.WriteString(linkContent)
		}
	}
	
	return strings.TrimSpace(content.String()), nil
}

func processAttachedFiles(attachFiles []string) (string, error) {
	var content strings.Builder
	
	content.WriteString("## Attached Files\n\n")
	
	for _, filePath := range attachFiles {
		log.Debug().Str("file", filePath).Msg("Processing attached file")
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Warn().Str("file", filePath).Msg("File does not exist, skipping")
			continue
		}
		
		// Read file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn().Str("file", filePath).Err(err).Msg("Failed to read file, skipping")
			continue
		}
		
		// Add file section
		content.WriteString(fmt.Sprintf("### %s\n\n", filepath.Base(filePath)))
		
		// Determine if it's a text file based on extension
		ext := strings.ToLower(filepath.Ext(filePath))
		textExts := map[string]bool{
			".txt": true, ".md": true, ".go": true, ".py": true, ".js": true,
			".ts": true, ".html": true, ".css": true, ".json": true, ".yaml": true,
			".yml": true, ".toml": true, ".xml": true, ".sql": true, ".sh": true,
			".bash": true, ".zsh": true, ".fish": true, ".conf": true, ".ini": true,
		}
		
		if textExts[ext] || ext == "" {
			// Text file - include content in code block
			language := getLanguageFromExtension(ext)
			content.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", language, string(fileContent)))
		} else {
			// Binary file - just mention it
			content.WriteString(fmt.Sprintf("*Binary file: %s (%d bytes)*\n\n", filePath, len(fileContent)))
		}
	}
	
	return content.String(), nil
}

func getLanguageFromExtension(ext string) string {
	langMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".html": "html",
		".css":  "css",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".toml": "toml",
		".xml":  "xml",
		".sql":  "sql",
		".sh":   "bash",
		".bash": "bash",
		".zsh":  "zsh",
		".fish": "fish",
		".md":   "markdown",
	}
	
	if lang, exists := langMap[ext]; exists {
		return lang
	}
	return ""
}

func askForLinks() ([]string, error) {
	var links []string
	
	fmt.Println("\nEnter relevant links (press Enter with empty line to finish):")
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("Link: ")
		if !scanner.Scan() {
			break
		}
		
		link := strings.TrimSpace(scanner.Text())
		if link == "" {
			break
		}
		
		links = append(links, link)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return links, nil
}

func processLinks(links []string) string {
	if len(links) == 0 {
		return ""
	}
	
	var content strings.Builder
	content.WriteString("## Links\n\n")
	
	for _, link := range links {
		link = strings.TrimSpace(link)
		if link != "" {
			// Simple URL validation - check if it starts with http/https
			if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", link, link))
			} else {
				// Assume it needs https:// prefix
				fullLink := "https://" + link
				content.WriteString(fmt.Sprintf("- [%s](%s)\n", link, fullLink))
			}
		}
	}
	
	content.WriteString("\n")
	return content.String()
}
