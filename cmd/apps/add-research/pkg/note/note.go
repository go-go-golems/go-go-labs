package note

import (
	"crypto/md5"
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
	"github.com/rs/zerolog/log"
)

type Config struct {
	VaultPath    string
	Title        string
	DateStr      string
	NoteType     string
	AppendMode   bool
	WithMetadata bool
}

type NoteInfo struct {
	Path      string
	Title     string
	Date      string
	Filename  string
	Size      int64
	WordCount int
	Preview   string
	ModTime   time.Time
}

type ExportConfig struct {
	VaultPath  string
	OutputPath string
	FromDate   string
	ToDate     string
}

func CreateNewNote(config Config, content string) error {
	log.Debug().Str("vaultPath", config.VaultPath).Msg("Creating new research note")

	var noteTitle string
	if config.Title != "" {
		// Use title from command line flag
		noteTitle = config.Title
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
	if config.DateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", config.DateStr)
		if err != nil {
			return errors.Wrap(err, "invalid date format, use YYYY-MM-DD")
		}
		targetDate = parsedDate
		log.Debug().Str("date", config.DateStr).Msg("Using date from command line flag")
	} else {
		targetDate = time.Now()
	}

	// Create date directory
	dateDir := filepath.Join(config.VaultPath, targetDate.Format("2006-01-02"))

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

	// Write file
	var fullContent string
	if config.WithMetadata {
		metadata := generateMetadata(noteTitle, targetDate, config.NoteType)
		fullContent = fmt.Sprintf("%s\n# %s\n\n%s\n", metadata, noteTitle, content)
	} else {
		fullContent = fmt.Sprintf("# %s\n\n%s\n", noteTitle, content)
	}

	if err := os.WriteFile(filepath_, []byte(fullContent), 0644); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	fmt.Printf("Created research note: %s\n", filepath_)
	return nil
}

func AppendToNote(config Config, content string) error {
	log.Debug().Str("vaultPath", config.VaultPath).Msg("Appending to research note")

	// Determine date to use
	var targetDate time.Time
	if config.DateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", config.DateStr)
		if err != nil {
			return errors.Wrap(err, "invalid date format, use YYYY-MM-DD")
		}
		targetDate = parsedDate
	} else {
		targetDate = time.Now()
	}

	dateDir := filepath.Join(config.VaultPath, targetDate.Format("2006-01-02"))

	// Check if date directory exists
	if _, err := os.Stat(dateDir); os.IsNotExist(err) {
		return errors.New("no research notes found for the specified date")
	}

	// Get existing files for the date
	files, err := getNotesForDate(dateDir)
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

func SearchNotes(vaultPath string) error {
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

				// Get file info
				fileInfo, err := d.Info()
				if err == nil {
					// Read content for word count and preview
					content, readErr := os.ReadFile(path)
					var wordCount int
					var preview string
					if readErr == nil {
						wordCount = countWords(string(content))
						preview = generatePreview(string(content))
					}

					allNotes = append(allNotes, NoteInfo{
						Path:      path,
						Title:     title,
						Date:      dateStr,
						Filename:  filename,
						Size:      fileInfo.Size(),
						WordCount: wordCount,
						Preview:   preview,
						ModTime:   fileInfo.ModTime(),
					})
				}
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
		displayName := fmt.Sprintf("%s - %s (%d words, %d bytes)", note.Date, note.Title, note.WordCount, note.Size)
		if note.Preview != "" {
			displayName += fmt.Sprintf("\n   %s", note.Preview)
		}
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

func getNotesForDate(dateDir string) ([]string, error) {
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

func generateMetadata(title string, date time.Time, noteType string) string {
	// Generate basic tags based on note type
	tags := []string{fmt.Sprintf("type/%s", noteType)}

	// Add year and month tags
	tags = append(tags, fmt.Sprintf("year/%d", date.Year()))
	tags = append(tags, fmt.Sprintf("month/%02d", date.Month()))

	// Generate unique ID/slug
	id := generateSlug(title)

	metadata := fmt.Sprintf(`---
title: "%s"
id: "%s"
slug: "%s"
date: %s
type: %s
tags:
  - %s
created: %s
modified: %s
source: "add-research-tool"
word_count: 0
---`,
		title,
		id,
		id,
		date.Format("2006-01-02"),
		noteType,
		strings.Join(tags, "\n  - "),
		date.Format(time.RFC3339),
		date.Format(time.RFC3339),
	)

	return metadata
}

func generateSlug(title string) string {
	// Simple slug generation: lowercase, replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	var result []rune
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}

	// Add timestamp-based hash for uniqueness
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%d", title, time.Now().UnixNano()))))
	return fmt.Sprintf("%s-%s", string(result), hash[:8])
}

func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func generatePreview(content string) string {
	// Remove markdown headers and get first few lines
	lines := strings.Split(content, "\n")
	var previewLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip YAML frontmatter, headers, and empty lines
		if line == "" || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "#") {
			continue
		}
		previewLines = append(previewLines, line)
		if len(previewLines) >= 2 {
			break
		}
	}

	preview := strings.Join(previewLines, " ")
	if len(preview) > 100 {
		// Truncate at word boundary
		runes := []rune(preview)
		if len(runes) > 100 {
			preview = string(runes[:100])
			lastSpace := strings.LastIndex(preview, " ")
			if lastSpace > 0 {
				preview = preview[:lastSpace]
			}
			preview += "..."
		}
	}

	return preview
}

func ExportNotes(config ExportConfig) error {
	log.Debug().Str("vaultPath", config.VaultPath).Msg("Exporting notes")

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = fmt.Sprintf("notes-export-%s.md", time.Now().Format("20060102"))
	}

	// Parse date range if provided
	var fromDate, toDate time.Time
	var err error
	if config.FromDate != "" {
		fromDate, err = time.Parse("2006-01-02", config.FromDate)
		if err != nil {
			return errors.Wrap(err, "invalid from date format, use YYYY-MM-DD")
		}
	}
	if config.ToDate != "" {
		toDate, err = time.Parse("2006-01-02", config.ToDate)
		if err != nil {
			return errors.Wrap(err, "invalid to date format, use YYYY-MM-DD")
		}
	}

	// Collect notes to export
	var notesToExport []NoteInfo

	err = filepath.WalkDir(config.VaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			relPath, err := filepath.Rel(config.VaultPath, path)
			if err != nil {
				return err
			}

			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 2 {
				dateStr := parts[0]

				// Check date range
				if config.FromDate != "" || config.ToDate != "" {
					noteDate, parseErr := time.Parse("2006-01-02", dateStr)
					if parseErr != nil {
						return nil // Skip notes with invalid date format
					}

					if config.FromDate != "" && noteDate.Before(fromDate) {
						return nil
					}
					if config.ToDate != "" && noteDate.After(toDate) {
						return nil
					}
				}

				filename := parts[1]
				title := filename
				if strings.Contains(filename, "-") {
					titleParts := strings.SplitN(filename, "-", 2)
					if len(titleParts) == 2 {
						title = strings.TrimSuffix(titleParts[1], ".md")
						title = strings.ReplaceAll(title, "-", " ")
					}
				}

				fileInfo, err := d.Info()
				if err == nil {
					notesToExport = append(notesToExport, NoteInfo{
						Path:     path,
						Title:    title,
						Date:     dateStr,
						Filename: filename,
						ModTime:  fileInfo.ModTime(),
					})
				}
			}
		}
		return nil
	})

	if err != nil {
		return errors.Wrap(err, "failed to collect notes for export")
	}

	if len(notesToExport) == 0 {
		fmt.Println("No notes found to export")
		return nil
	}

	// Sort by date (oldest first for export)
	sort.Slice(notesToExport, func(i, j int) bool {
		return notesToExport[i].Date < notesToExport[j].Date
	})

	// Create export file
	exportFile, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrap(err, "failed to create export file")
	}
	defer exportFile.Close()

	// Write header
	fmt.Fprintf(exportFile, "# Research Notes Export\n\n")
	fmt.Fprintf(exportFile, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	if config.FromDate != "" || config.ToDate != "" {
		fmt.Fprintf(exportFile, "Date range: %s to %s\n", config.FromDate, config.ToDate)
	}
	fmt.Fprintf(exportFile, "Total notes: %d\n\n", len(notesToExport))
	fmt.Fprintf(exportFile, "---\n\n")

	// Export each note
	for _, note := range notesToExport {
		content, err := os.ReadFile(note.Path)
		if err != nil {
			log.Warn().Str("path", note.Path).Err(err).Msg("Failed to read note, skipping")
			continue
		}

		fmt.Fprintf(exportFile, "## %s (%s)\n\n", note.Title, note.Date)
		fmt.Fprintf(exportFile, "Path: `%s`\n\n", note.Path)
		fmt.Fprintf(exportFile, "%s\n\n", string(content))
		fmt.Fprintf(exportFile, "---\n\n")
	}

	fmt.Printf("Exported %d notes to: %s\n", len(notesToExport), outputPath)
	return nil
}
