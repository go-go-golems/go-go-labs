package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	appendMode  bool
	logLevel    string
	message     string
	attachFiles []string
	useClipboard bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "add-research",
		Short: "Add research notes to obsidian vault",
		Long:  "Interactive tool to create or append to research notes in your obsidian vault",
		RunE:  runCommand,
	}

	rootCmd.Flags().BoolVar(&appendMode, "append", false, "Append to existing research note")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVarP(&message, "message", "m", "", "Research note content from command line")
	rootCmd.Flags().StringSliceVarP(&attachFiles, "file", "f", []string{}, "Files to attach to the research note")
	rootCmd.Flags().BoolVarP(&useClipboard, "clip", "c", false, "Use content from clipboard")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	setupLogging()

	vaultPath := filepath.Join(os.Getenv("HOME"), "code", "wesen", "obsidian-vault", "Research")
	
	if appendMode {
		return appendToResearch(vaultPath)
	}
	
	return createNewResearch(vaultPath)
}

func setupLogging() {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func createNewResearch(vaultPath string) error {
	log.Debug().Str("vaultPath", vaultPath).Msg("Creating new research note")
	
	var title string
	err := huh.NewInput().
		Title("Research Note Title").
		Placeholder("Enter title for your research note...").
		Value(&title).
		Run()
	if err != nil {
		return errors.Wrap(err, "failed to get title")
	}

	if strings.TrimSpace(title) == "" {
		return errors.New("title cannot be empty")
	}

	// Create date directory
	today := time.Now().Format("2006-01-02")
	dateDir := filepath.Join(vaultPath, today)
	
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create date directory")
	}

	// Find next incremental number
	nextNum, err := getNextIncrementalNumber(dateDir)
	if err != nil {
		return errors.Wrap(err, "failed to get next incremental number")
	}

	// Create filename
	filename := fmt.Sprintf("%03d-%s.md", nextNum, sanitizeFilename(title))
	filepath_ := filepath.Join(dateDir, filename)

	log.Info().Str("filepath", filepath_).Msg("Creating research note")

	// Get content from user
	content, err := getContentFromUser()
	if err != nil {
		return errors.Wrap(err, "failed to get content")
	}

	// Write file
	fullContent := fmt.Sprintf("# %s\n\n%s\n", title, content)
	if err := os.WriteFile(filepath_, []byte(fullContent), 0644); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	fmt.Printf("Created research note: %s\n", filepath_)
	return nil
}

func appendToResearch(vaultPath string) error {
	log.Debug().Str("vaultPath", vaultPath).Msg("Appending to research note")
	
	today := time.Now().Format("2006-01-02")
	dateDir := filepath.Join(vaultPath, today)

	// Check if date directory exists
	if _, err := os.Stat(dateDir); os.IsNotExist(err) {
		return errors.New("no research notes found for today")
	}

	// Get existing files for today
	files, err := getResearchFilesForDate(dateDir)
	if err != nil {
		return errors.Wrap(err, "failed to get research files")
	}

	if len(files) == 0 {
		return errors.New("no research notes found for today")
	}

	// Let user select file
	var selectedFile string
	options := make([]huh.Option[string], len(files))
	for i, file := range files {
		displayName := strings.TrimSuffix(filepath.Base(file), ".md")
		options[i] = huh.NewOption(displayName, file)
	}

	err = huh.NewSelect[string]().
		Title("Select research note to append to").
		Options(options...).
		Value(&selectedFile).
		Run()
	if err != nil {
		return errors.Wrap(err, "failed to select file")
	}

	// Get content from user
	content, err := getContentFromUser()
	if err != nil {
		return errors.Wrap(err, "failed to get content")
	}

	// Append to file
	f, err := os.OpenFile(selectedFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open file for appending")
	}
	defer f.Close()

	appendContent := fmt.Sprintf("\n---\n\n%s\n", content)
	if _, err := f.WriteString(appendContent); err != nil {
		return errors.Wrap(err, "failed to append content")
	}

	fmt.Printf("Appended to research note: %s\n", selectedFile)
	return nil
}

func getNextIncrementalNumber(dateDir string) (int, error) {
	files, err := os.ReadDir(dateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxNum := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		parts := strings.SplitN(file.Name(), "-", 2)
		if len(parts) < 2 {
			continue
		}
		
		if num, err := strconv.Atoi(parts[0]); err == nil && num > maxNum {
			maxNum = num
		}
	}

	return maxNum + 1, nil
}

func getResearchFilesForDate(dateDir string) ([]string, error) {
	var files []string
	
	err := filepath.WalkDir(dateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort files by name (which includes the incremental number)
	sort.Strings(files)
	return files, nil
}

func sanitizeFilename(filename string) string {
	// Replace spaces with dashes and remove invalid characters
	filename = strings.ReplaceAll(filename, " ", "-")
	filename = strings.ReplaceAll(filename, "/", "-")
	filename = strings.ReplaceAll(filename, "\\", "-")
	filename = strings.ReplaceAll(filename, ":", "-")
	filename = strings.ReplaceAll(filename, "*", "-")
	filename = strings.ReplaceAll(filename, "?", "-")
	filename = strings.ReplaceAll(filename, "\"", "-")
	filename = strings.ReplaceAll(filename, "<", "-")
	filename = strings.ReplaceAll(filename, ">", "-")
	filename = strings.ReplaceAll(filename, "|", "-")
	
	return filename
}

func getContentFromUser() (string, error) {
	var content strings.Builder
	
	// If message provided via command line, use it
	if message != "" {
		content.WriteString(message)
		content.WriteString("\n")
	}
	
	// If clipboard flag is set, read from clipboard
	if useClipboard {
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
	} else if message == "" && !useClipboard {
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
	if len(attachFiles) > 0 {
		fileContent, err := processAttachedFiles()
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
	
	return strings.TrimSpace(content.String()), nil
}

func processAttachedFiles() (string, error) {
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
