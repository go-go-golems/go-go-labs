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
	appendMode   bool
	logLevel     string
	message      string
	attachFiles  []string
	useClipboard bool
	title        string
	dateStr      string
	noteType     string
	searchMode   bool
	browseFiles  bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "add-research",
		Short: "Add research notes to obsidian vault",
		Long:  "Interactive tool to create, search, or append to research notes in your obsidian vault",
		RunE:  runCommand,
	}

	rootCmd.Flags().BoolVar(&appendMode, "append", false, "Append to existing research note")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVarP(&message, "message", "m", "", "Research note content from command line")
	rootCmd.Flags().StringSliceVarP(&attachFiles, "file", "f", []string{}, "Files to attach to the research note")
	rootCmd.Flags().BoolVarP(&useClipboard, "clip", "c", false, "Use content from clipboard")
	rootCmd.Flags().StringVarP(&title, "title", "t", "", "Title for the research note (skips interactive input)")
	rootCmd.Flags().StringVar(&dateStr, "date", "", "Date for the note (YYYY-MM-DD format, default today)")
	rootCmd.Flags().StringVar(&noteType, "type", "research", "Type of note (research, ideas, notes, etc)")
	rootCmd.Flags().BoolVarP(&searchMode, "search", "s", false, "Search existing notes by name (fuzzy)")
	rootCmd.Flags().BoolVarP(&browseFiles, "browse", "b", false, "Browse and select files to attach interactively")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	setupLogging()

	// Handle file browsing first if requested
	if browseFiles {
		selectedFiles, err := browseForFiles()
		if err != nil {
			return errors.Wrap(err, "failed to browse files")
		}
		attachFiles = append(attachFiles, selectedFiles...)
		log.Debug().Strs("selectedFiles", selectedFiles).Msg("Added files from browser")
	}

	vaultPath := filepath.Join(os.Getenv("HOME"), "code", "wesen", "obsidian-vault", noteType)
	
	if searchMode {
		return searchNotes(vaultPath)
	}
	
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
	
	var noteTitle string
	if title != "" {
		// Use title from command line flag
		noteTitle = title
		log.Debug().Str("title", noteTitle).Msg("Using title from command line flag")
	} else {
		// Interactive title input
		err := huh.NewInput().
			Title("Research Note Title").
			Placeholder("Enter title for your research note...").
			Value(&noteTitle).
			Run()
		if err != nil {
			return errors.Wrap(err, "failed to get title")
		}
	}

	if strings.TrimSpace(noteTitle) == "" {
		return errors.New("title cannot be empty")
	}

	// Determine date to use
	var targetDate time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return errors.Wrap(err, "invalid date format, use YYYY-MM-DD")
		}
		targetDate = parsedDate
		log.Debug().Str("date", dateStr).Msg("Using date from command line flag")
	} else {
		targetDate = time.Now()
	}

	// Create date directory
	dateDir := filepath.Join(vaultPath, targetDate.Format("2006-01-02"))
	
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create date directory")
	}

	// Find next incremental number
	nextNum, err := getNextIncrementalNumber(dateDir)
	if err != nil {
		return errors.Wrap(err, "failed to get next incremental number")
	}

	// Create filename
	filename := fmt.Sprintf("%03d-%s.md", nextNum, sanitizeFilename(noteTitle))
	filepath_ := filepath.Join(dateDir, filename)

	log.Info().Str("filepath", filepath_).Msg("Creating research note")

	// Get content from user
	content, err := getContentFromUser()
	if err != nil {
		return errors.Wrap(err, "failed to get content")
	}

	// Write file
	fullContent := fmt.Sprintf("# %s\n\n%s\n", noteTitle, content)
	if err := os.WriteFile(filepath_, []byte(fullContent), 0644); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	fmt.Printf("Created research note: %s\n", filepath_)
	return nil
}

func appendToResearch(vaultPath string) error {
	log.Debug().Str("vaultPath", vaultPath).Msg("Appending to research note")
	
	// Determine date to use
	var targetDate time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return errors.Wrap(err, "invalid date format, use YYYY-MM-DD")
		}
		targetDate = parsedDate
	} else {
		targetDate = time.Now()
	}
	
	dateDir := filepath.Join(vaultPath, targetDate.Format("2006-01-02"))

	// Check if date directory exists
	if _, err := os.Stat(dateDir); os.IsNotExist(err) {
		return errors.New("no research notes found for the specified date")
	}

	// Get existing files for today
	files, err := getResearchFilesForDate(dateDir)
	if err != nil {
		return errors.Wrap(err, "failed to get research files")
	}

	if len(files) == 0 {
		return errors.New("no research notes found for the specified date")
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

func searchNotes(vaultPath string) error {
	log.Debug().Str("vaultPath", vaultPath).Msg("Searching notes")
	
	// Get all notes
	var allNotes []NoteInfo
	
	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			// Parse note info
			relPath, err := filepath.Rel(vaultPath, path)
			if err != nil {
				return err
			}
			
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 2 {
				dateStr := parts[0]
				filename := parts[1]
				
				// Extract title from filename (remove number prefix and .md suffix)
				title := filename
				if strings.Contains(filename, "-") {
					titleParts := strings.SplitN(filename, "-", 2)
					if len(titleParts) == 2 {
						title = strings.TrimSuffix(titleParts[1], ".md")
						title = strings.ReplaceAll(title, "-", " ")
					}
				}
				
				allNotes = append(allNotes, NoteInfo{
					Path:     path,
					Title:    title,
					Date:     dateStr,
					Filename: filename,
				})
			}
		}
		return nil
	})
	
	if err != nil {
		return errors.Wrap(err, "failed to search notes")
	}
	
	if len(allNotes) == 0 {
		fmt.Println("No notes found")
		return nil
	}
	
	// Sort by date (newest first)
	sort.Slice(allNotes, func(i, j int) bool {
		return allNotes[i].Date > allNotes[j].Date
	})
	
	// Interactive selection
	var selectedNote NoteInfo
	options := make([]huh.Option[NoteInfo], len(allNotes))
	for i, note := range allNotes {
		displayName := fmt.Sprintf("%s - %s", note.Date, note.Title)
		options[i] = huh.NewOption(displayName, note)
	}
	
	err = huh.NewSelect[NoteInfo]().
		Title("Select a note").
		Options(options...).
		Value(&selectedNote).
		Run()
	if err != nil {
		return errors.Wrap(err, "failed to select note")
	}
	
	// Read and display the note
	content, err := os.ReadFile(selectedNote.Path)
	if err != nil {
		return errors.Wrap(err, "failed to read note")
	}
	
	fmt.Printf("=== %s ===\n", selectedNote.Title)
	fmt.Printf("Date: %s\n", selectedNote.Date)
	fmt.Printf("Path: %s\n\n", selectedNote.Path)
	fmt.Print(string(content))
	
	// Ask if user wants to copy to clipboard
	var copyToClip bool
	err = huh.NewConfirm().
		Title("Copy content to clipboard?").
		Value(&copyToClip).
		Run()
	if err != nil {
		return errors.Wrap(err, "failed to get clipboard confirmation")
	}
	
	if copyToClip {
		err = clipboard.WriteAll(string(content))
		if err != nil {
			return errors.Wrap(err, "failed to copy to clipboard")
		}
		fmt.Println("Content copied to clipboard!")
	}
	
	return nil
}

type NoteInfo struct {
	Path     string
	Title    string
	Date     string
	Filename string
}

func browseForFiles() ([]string, error) {
	var selectedFiles []string
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current directory")
	}

	for {
		// List files and directories in current directory
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read directory")
		}

		// Create options for selection
		var options []huh.Option[string]
		
		// Add parent directory option (unless we're at root)
		if currentDir != "/" {
			options = append(options, huh.NewOption(".. (parent directory)", ".."))
		}

		// Add directories first
		for _, entry := range entries {
			if entry.IsDir() {
				displayName := fmt.Sprintf("📁 %s/", entry.Name())
				options = append(options, huh.NewOption(displayName, entry.Name()))
			}
		}

		// Add files
		for _, entry := range entries {
			if !entry.IsDir() {
				displayName := fmt.Sprintf("📄 %s", entry.Name())
				options = append(options, huh.NewOption(displayName, entry.Name()))
			}
		}

		// Add special options
		options = append(options, huh.NewOption("✅ Done selecting", "DONE"))
		options = append(options, huh.NewOption("❌ Cancel", "CANCEL"))

		var selection string
		err = huh.NewSelect[string]().
			Title(fmt.Sprintf("Browse files in: %s", currentDir)).
			Description(fmt.Sprintf("Selected files: %d", len(selectedFiles))).
			Options(options...).
			Value(&selection).
			Run()
		if err != nil {
			return nil, errors.Wrap(err, "failed to select item")
		}

		switch selection {
		case "DONE":
			return selectedFiles, nil
		case "CANCEL":
			return []string{}, nil
		case "..":
			currentDir = filepath.Dir(currentDir)
		default:
			targetPath := filepath.Join(currentDir, selection)
			stat, err := os.Stat(targetPath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to stat selected item")
			}

			if stat.IsDir() {
				currentDir = targetPath
			} else {
				// It's a file, add it to selected files
				selectedFiles = append(selectedFiles, targetPath)
				log.Debug().Str("file", targetPath).Msg("Added file to selection")
			}
		}
	}
}
