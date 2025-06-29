package main

import (
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-labs/cmd/apps/add-research/pkg/browser"
	"github.com/go-go-golems/go-go-labs/cmd/apps/add-research/pkg/content"
	"github.com/go-go-golems/go-go-labs/cmd/apps/add-research/pkg/note"
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
	askForLinks  bool
	linksSlice   []string
	withMetadata bool
	noLinks      bool
	exportMode   bool
	exportPath   string
	exportFrom   string
	exportTo     string
	configPath   string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "add-research",
		Short: "Add research notes to obsidian vault",
		Long: `Interactive tool to create, search, export, or append to research notes in your obsidian vault.

By default, the tool will ask for links interactively. Use --links to provide them via command line,
or --no-links to disable link prompting entirely.

Examples:
  add-research                           # Create new note with interactive link input
  add-research --links "https://example.com" "https://github.com/user/repo"
  add-research --no-links               # Create note without any links
  add-research --search                 # Search existing notes (shows previews and metadata)
  add-research --export --export-from "2024-01-01" --export-to "2024-12-31"
  add-research --append --date "2024-01-15"`,
		RunE: runCommand,
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
	rootCmd.Flags().BoolVar(&askForLinks, "ask-links", false, "Prompt for relevant links to include in the note (deprecated - now default)")
	rootCmd.Flags().StringSliceVar(&linksSlice, "links", []string{}, "Links to include in the note (skips interactive prompting)")
	rootCmd.Flags().BoolVar(&noLinks, "no-links", false, "Disable link prompting entirely")
	rootCmd.Flags().BoolVar(&withMetadata, "metadata", false, "Include YAML frontmatter with metadata")
	rootCmd.Flags().BoolVar(&exportMode, "export", false, "Export notes to a combined markdown file")
	rootCmd.Flags().StringVar(&exportPath, "export-path", "", "Path for exported file (default: notes-export-YYYYMMDD.md)")
	rootCmd.Flags().StringVar(&exportFrom, "export-from", "", "Export from date (YYYY-MM-DD)")
	rootCmd.Flags().StringVar(&exportTo, "export-to", "", "Export to date (YYYY-MM-DD)")
	rootCmd.Flags().StringVar(&configPath, "config", "", "Config file path (default: ~/.add-research.yaml)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	setupLogging()

	// Load config from file if specified or from default location
	config, err := loadConfig()
	if err != nil {
		log.Debug().Err(err).Msg("Could not load config file, using defaults")
	}

	// Apply config defaults if values not set via flags
	applyConfigDefaults(config)

	// Handle file browsing first if requested
	if browseFiles {
		selectedFiles, err := browser.BrowseForFiles()
		if err != nil {
			return errors.Wrap(err, "failed to browse files")
		}
		attachFiles = append(attachFiles, selectedFiles...)
		log.Debug().Strs("selectedFiles", selectedFiles).Msg("Added files from browser")
	}

	vaultPath := filepath.Join(os.Getenv("HOME"), "code", "wesen", "obsidian-vault", noteType)
	
	if exportMode {
		return note.ExportNotes(note.ExportConfig{
			VaultPath:  vaultPath,
			OutputPath: exportPath,
			FromDate:   exportFrom,
			ToDate:     exportTo,
		})
	}
	
	if searchMode {
		return note.SearchNotes(vaultPath)
	}
	
	// Determine if we should ask for links
	shouldAskForLinks := determineLinkBehavior()
	
	// Get content from user for create/append operations
	contentConfig := content.Config{
		Message:      message,
		UseClipboard: useClipboard,
		AttachFiles:  attachFiles,
		AskForLinks:  shouldAskForLinks,
		Links:        linksSlice,
	}
	
	noteContent, err := content.GetContentFromUser(contentConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get content")
	}
	
	noteConfig := note.Config{
		VaultPath:    vaultPath,
		Title:        title,
		DateStr:      dateStr,
		NoteType:     noteType,
		AppendMode:   appendMode,
		WithMetadata: withMetadata,
	}
	
	if appendMode {
		return note.AppendToNote(noteConfig, noteContent)
	}
	
	return note.CreateNewNote(noteConfig, noteContent)
}

func setupLogging() {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

type AppConfig struct {
	VaultBasePath    string `yaml:"vault_base_path"`
	DefaultNoteType  string `yaml:"default_note_type"`
	WithMetadata     bool   `yaml:"with_metadata"`
	AskForLinks      bool   `yaml:"ask_for_links"`
}

func loadConfig() (*AppConfig, error) {
	configFile := configPath
	if configFile == "" {
		home := os.Getenv("HOME")
		configFile = filepath.Join(home, ".add-research.yaml")
	}
	
	_, err := os.ReadFile(configFile)
	if err != nil {
		return &AppConfig{}, err
	}
	
	var config AppConfig
	// Note: We'd need to import gopkg.in/yaml.v3 for this to work
	// For now, return empty config
	log.Debug().Str("configFile", configFile).Msg("Config file found but yaml parsing not implemented")
	return &config, nil
}

func applyConfigDefaults(config *AppConfig) {
	// Apply config defaults only if flags weren't explicitly set
	// This is a simplified implementation - in production you'd want to check
	// if flags were actually set vs default values
}

func determineLinkBehavior() bool {
	// Priority: --no-links > --links provided > default (ask for links)
	if noLinks {
		log.Debug().Msg("Links disabled via --no-links flag")
		return false
	}
	
	if len(linksSlice) > 0 {
		log.Debug().Strs("links", linksSlice).Msg("Using provided links, skipping interactive input")
		return false
	}
	
	// Default behavior: ask for links
	log.Debug().Msg("Using default behavior: asking for links interactively")
	return true
}
