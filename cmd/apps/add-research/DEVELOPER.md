# 🛠️ Developer Guide - Add Research Tool

> Comprehensive developer documentation for understanding, modifying, and extending the add-research CLI tool

## 📋 Project Overview

### Architecture Goals

The add-research tool follows a clean, modular architecture designed for:

- **Simplicity**: Easy to understand and modify
- **Extensibility**: New features can be added without breaking existing functionality  
- **Testability**: Clear separation of concerns enables comprehensive testing
- **CLI-First Design**: Built around cobra for excellent command-line UX
- **Obsidian Integration**: Seamless workflow with existing note-taking systems

### Tech Stack

- **Language**: Go 1.23+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) for command parsing
- **TUI Components**: [Huh](https://github.com/charmbracelet/huh) for interactive forms
- **Configuration**: YAML with Viper (planned)
- **Logging**: [Zerolog](https://github.com/rs/zerolog) for structured logging
- **Error Handling**: [pkg/errors](https://github.com/pkg/errors) for error wrapping
- **Clipboard**: [atotto/clipboard](https://github.com/atotto/clipboard) for system clipboard integration

### Project Structure

```
cmd/apps/add-research/
├── main.go              # Entry point, CLI setup, command orchestration
├── pkg/
│   ├── note/            # Core note operations (create, search, export)
│   ├── content/         # Content gathering from various sources
│   └── browser/         # Interactive file browser
├── README.md            # User documentation
├── DEVELOPER.md         # This file
└── TODO.md              # Feature roadmap
```

## 🗺️ Codebase Tour

### 📄 [`main.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/main.go) - Application Entry Point

The main file orchestrates the entire application flow:

**Key Responsibilities**:
- Command-line flag parsing and validation
- Logging setup and configuration loading
- Mode detection (create/append/search/export)
- High-level workflow coordination

**Important Functions**:
- `main()` - Sets up cobra command with all flags
- `runCommand()` - Main application logic and flow control  
- `setupLogging()` - Configures zerolog based on user preferences
- `determineLinkBehavior()` - Implements link collection priority logic

**Code Flow Pseudocode**:
```go
func runCommand() {
    setupLogging()
    config := loadConfig()
    
    if browseFiles {
        files := browser.BrowseForFiles()
        attachFiles = append(attachFiles, files...)
    }
    
    vaultPath := determineVaultPath()
    
    switch mode {
    case export:
        return note.ExportNotes(exportConfig)
    case search:
        return note.SearchNotes(vaultPath)  
    case append:
        content := content.GetContentFromUser(contentConfig)
        return note.AppendToNote(noteConfig, content)
    default: // create
        content := content.GetContentFromUser(contentConfig)
        return note.CreateNewNote(noteConfig, content)
    }
}
```

### 📝 [`pkg/note/note.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/note/note.go) - Core Note Operations

This package handles all note-related operations and is the heart of the application:

**Package Purpose**: 
Manages the complete lifecycle of research notes including creation, modification, searching, and exporting. Implements the core business logic for organizing notes in a date-based hierarchy with automatic numbering.

**Key Data Structures**:
- `Config` - Configuration for note operations (vault path, title, date, etc.)
- `NoteInfo` - Rich metadata about existing notes (path, word count, preview, etc.)
- `ExportConfig` - Settings for exporting note collections

**Core Functions**:

- **`CreateNewNote(config Config, content string) error`** - [Lines 47-118](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/note/note.go#L47-L118)
  - Handles title input (command-line or interactive)
  - Creates date-based directory structure (`vault/type/YYYY-MM-DD/`)
  - Generates incremental numbering (`001-title.md`, `002-title.md`, etc.)
  - Writes markdown content with optional YAML frontmatter

- **`AppendToNote(config Config, content string) error`** - [Lines 120-183](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/note/note.go#L120-L183)  
  - Finds existing notes for specified date
  - Presents interactive selection menu
  - Appends content with separator (`---`)

- **`SearchNotes(vaultPath string) error`** - [Lines 185-310](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/note/note.go#L185-L310)
  - Recursively scans vault directory
  - Extracts metadata (word count, file size, modification time)
  - Generates content previews
  - Provides interactive selection with clipboard integration

- **`ExportNotes(config ExportConfig) error`** - [Lines 473-602](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/note/note.go#L473-L602)
  - Filters notes by date range
  - Combines multiple notes into single markdown file
  - Includes metadata headers and file paths

**Helper Functions**:
- `getNextIncrementalNumber()` - Scans directory for highest numbered file
- `sanitizeFilename()` - Cleans titles for safe filesystem use
- `generateMetadata()` - Creates YAML frontmatter with auto-generated IDs
- `generateSlug()` - Creates URL-safe identifiers with MD5 uniqueness
- `countWords()` and `generatePreview()` - Content analysis for search results

**File Organization Logic**:
```
vault/
└── research/                    # Note type
    └── 2024-01-15/             # Date directory  
        ├── 001-first-note.md   # Incremental numbering
        ├── 002-second-note.md
        └── 003-third-note.md
```

### 📄 [`pkg/content/content.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/content/content.go) - Content Gathering

This package handles collecting content from multiple sources and formatting it for notes:

**Package Purpose**:
Abstracts the complexity of gathering content from various input sources (command-line, clipboard, stdin, files, interactive input) and formats it consistently for note creation.

**Key Functions**:

- **`GetContentFromUser(config Config) (string, error)`** - [Lines 23-109](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/content/content.go#L23-L109)
  - Orchestrates content collection from multiple sources
  - Handles priority: command-line message → clipboard → stdin → interactive input
  - Processes attached files and links
  - Combines everything into structured markdown

**Content Source Priority**:
```go
priority := []source{
    commandLineMessage,  // --message flag
    clipboardContent,    // --clip flag  
    stdinPipe,          // echo "content" | add-research
    interactiveInput,   // Terminal prompt
}
```

- **`processAttachedFiles(attachFiles []string) (string, error)`** - [Lines 111-155](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/content/content.go#L111-L155)
  - Reads and embeds file contents
  - Detects file types by extension
  - Formats text files with appropriate syntax highlighting
  - Handles binary files with metadata only

- **`askForLinks() ([]string, error)`** - [Lines 184-209](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/content/content.go#L184-L209)
  - Interactive link collection
  - Continues until empty input
  - Basic URL validation and formatting

**Supported File Types**:
Text files get embedded with syntax highlighting:
- Code: `.go`, `.py`, `.js`, `.ts`, `.html`, `.css`
- Data: `.json`, `.yaml`, `.xml`, `.sql`  
- Scripts: `.sh`, `.bash`, `.zsh`
- Documentation: `.md`, `.txt`

Binary files get referenced with metadata only.

### 🗂️ [`pkg/browser/browser.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/browser/browser.go) - Interactive File Browser

Simple but effective file browser for selecting multiple files:

**Package Purpose**:
Provides a terminal-based file browser interface for selecting files to attach to notes. Uses the Huh TUI library for an intuitive navigation experience.

**Key Function**:
- **`BrowseForFiles() ([]string, error)`** - [Lines 13-89](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/add-research/pkg/browser/browser.go#L13-L89)

**Browser Features**:
- Navigation: Parent directory (`..`), subdirectories, files
- Visual indicators: 📁 for directories, 📄 for files
- Multi-selection: Add multiple files before finishing
- Clean exit: "Done" and "Cancel" options

**Navigation Flow**:
```
Current: /home/user/projects
📁 src/
📁 docs/  
📄 README.md
📄 config.yaml
✅ Done selecting (2 files selected)
❌ Cancel
```

## 🏗️ Development Setup

### Prerequisites

```bash
# Go installation
go version  # Should show 1.23+

# Clone repository
git clone https://github.com/go-go-golems/go-go-labs.git
cd go-go-labs/cmd/apps/add-research
```

### Development Workflow

```bash
# Install dependencies
go mod tidy

# Run in development mode
go run main.go --help

# Run with debug logging
go run main.go --log-level debug --title "Test Note" --message "Testing"

# Build binary
go build -o add-research-dev main.go

# Test the binary
./add-research-dev --version
```

### Environment Setup

The tool expects this directory structure:
```bash
# Default vault location (can be customized)
export VAULT_PATH="$HOME/code/wesen/obsidian-vault"
mkdir -p "$VAULT_PATH/research"
mkdir -p "$VAULT_PATH/ideas" 
mkdir -p "$VAULT_PATH/notes"
```

## 🏛️ Architecture Deep Dive

### Package Responsibilities

```
┌─────────────────────────────────────────────────────┐
│                    main.go                          │
│  • CLI parsing & validation                         │
│  • Configuration loading                            │  
│  • High-level workflow orchestration                │
│  • Error handling & logging setup                   │
└─────────────────┬───────────────────────────────────┘
                  │
    ┌─────────────┼─────────────┬─────────────────────┐
    │             │             │                     │
    ▼             ▼             ▼                     ▼
┌─────────┐  ┌─────────┐  ┌──────────┐  ┌────────────┐
│ content │  │  note   │  │ browser  │  │   config   │
│ package │  │ package │  │ package  │  │ (future)   │
└─────────┘  └─────────┘  └──────────┘  └────────────┘
```

### Data Flow Diagram

```
User Input Sources:
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Command     │    │ Clipboard   │    │ Interactive │
│ Line Args   │    │ Content     │    │ Prompts     │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                          ▼
                ┌─────────────────┐
                │ content.Config  │
                │ & Processor     │
                └─────────┬───────┘
                          │
                          ▼
                ┌─────────────────┐
                │ Formatted       │
                │ Markdown        │
                │ Content         │
                └─────────┬───────┘
                          │
                          ▼
                ┌─────────────────┐
                │ note.Config     │
                │ & Operations    │
                └─────────┬───────┘
                          │
                          ▼
                ┌─────────────────┐
                │ File System     │
                │ vault/type/     │
                │ YYYY-MM-DD/     │
                │ NNN-title.md    │
                └─────────────────┘
```

### Key Design Decisions

**1. Date-Based Organization**
- **Decision**: Use `YYYY-MM-DD` directories with incremental numbering
- **Rationale**: Natural chronological browsing, prevents filename conflicts
- **Trade-offs**: Slightly more complex search, but much better organization

**2. Incremental Numbering**  
- **Decision**: `001-title.md`, `002-title.md` format
- **Rationale**: Maintains creation order, handles duplicate titles
- **Implementation**: Scan directory for highest number, increment by 1

**3. Modular Package Structure**
- **Decision**: Separate packages for note, content, browser operations
- **Rationale**: Clear separation of concerns, easier testing, reusable components
- **Benefits**: Each package can be tested independently

**4. Interactive TUI Components**
- **Decision**: Use Huh library for forms and selection
- **Rationale**: Better UX than raw terminal input, consistent styling
- **Trade-offs**: Additional dependency, but significantly better user experience

**5. Multiple Content Sources**
- **Decision**: Support command-line, clipboard, stdin, interactive, and file sources
- **Rationale**: Flexibility for different workflows and automation scenarios
- **Implementation**: Priority-based processing with clear fallback chain

### Error Handling Strategy

```go
// Pattern used throughout codebase
func SomeOperation() error {
    result, err := riskyOperation()
    if err != nil {
        return errors.Wrap(err, "contextual description of what failed")
    }
    
    // ... continue processing
    return nil
}
```

**Benefits**:
- Clear error context for debugging
- Structured error messages for users  
- Easy to trace error origins in logs

## 🔧 Adding New Features

### Step-by-Step Guide with Examples

#### Example 1: Adding Tag Support

**1. Update Data Structures**

```go
// In pkg/note/note.go
type Config struct {
    VaultPath    string
    Title        string
    DateStr      string
    NoteType     string
    AppendMode   bool
    WithMetadata bool
    Tags         []string  // NEW FIELD
}
```

**2. Add CLI Flag**

```go
// In main.go
var tags []string  // NEW VARIABLE

func main() {
    // ... existing flags ...
    rootCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Tags to add to the note")
}
```

**3. Pass Through Configuration**

```go
// In main.go runCommand()
noteConfig := note.Config{
    // ... existing fields ...
    Tags: tags,  // NEW FIELD
}
```

**4. Implement Feature Logic**

```go
// In pkg/note/note.go generateMetadata()
func generateMetadata(title string, date time.Time, noteType string, userTags []string) string {
    tags := []string{fmt.Sprintf("type/%s", noteType)}
    tags = append(tags, userTags...)  // Add user-provided tags
    
    // ... rest of function
}
```

**5. Update Function Calls**

```go
// In pkg/note/note.go CreateNewNote()
if config.WithMetadata {
    metadata := generateMetadata(noteTitle, targetDate, config.NoteType, config.Tags)
    // ...
}
```

#### Example 2: Adding Template Support

**1. Create New Package**

```bash
mkdir pkg/template
touch pkg/template/template.go
```

**2. Define Template Structure**

```go
// pkg/template/template.go
package template

import (
    "path/filepath"
    "text/template"
)

type TemplateConfig struct {
    TemplatePath string
    Variables    map[string]interface{}
}

func ApplyTemplate(config TemplateConfig, content string) (string, error) {
    if config.TemplatePath == "" {
        return content, nil  // No template, return original
    }
    
    tmpl, err := template.ParseFiles(config.TemplatePath)
    if err != nil {
        return "", errors.Wrap(err, "failed to parse template")
    }
    
    var result strings.Builder
    err = tmpl.Execute(&result, config.Variables)
    if err != nil {
        return "", errors.Wrap(err, "failed to execute template")
    }
    
    return result.String(), nil
}
```

**3. Integrate with Existing Flow**

```go
// In pkg/note/note.go CreateNewNote()
import "github.com/go-go-golems/go-go-labs/cmd/apps/add-research/pkg/template"

func CreateNewNote(config Config, content string) error {
    // ... existing code ...
    
    // Apply template if configured
    if config.TemplatePath != "" {
        templateConfig := template.TemplateConfig{
            TemplatePath: config.TemplatePath,
            Variables: map[string]interface{}{
                "Title": noteTitle,
                "Date":  targetDate.Format("2006-01-02"),
                "Type":  config.NoteType,
            },
        }
        
        content, err = template.ApplyTemplate(templateConfig, content)
        if err != nil {
            return errors.Wrap(err, "failed to apply template")
        }
    }
    
    // ... rest of function
}
```

## 🧪 Testing Strategy

### Test Structure

```bash
# Test organization
pkg/
├── note/
│   ├── note.go
│   └── note_test.go
├── content/
│   ├── content.go  
│   └── content_test.go
└── browser/
    ├── browser.go
    └── browser_test.go
```

### Unit Testing Examples

```go
// pkg/note/note_test.go
package note

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestCreateNewNote(t *testing.T) {
    // Setup temporary vault
    tempDir := t.TempDir()
    
    config := Config{
        VaultPath:    tempDir,
        Title:        "Test Note",
        DateStr:      "2024-01-15",
        NoteType:     "research",
        AppendMode:   false,
        WithMetadata: true,
    }
    
    content := "This is test content"
    
    err := CreateNewNote(config, content)
    if err != nil {
        t.Fatalf("CreateNewNote failed: %v", err)
    }
    
    // Verify file was created
    expectedPath := filepath.Join(tempDir, "2024-01-15", "001-Test-Note.md")
    if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
        t.Errorf("Expected file %s was not created", expectedPath)
    }
    
    // Verify content
    fileContent, err := os.ReadFile(expectedPath)
    if err != nil {
        t.Fatalf("Failed to read created file: %v", err)
    }
    
    contentStr := string(fileContent)
    if !strings.Contains(contentStr, "Test Note") {
        t.Errorf("File content missing title")
    }
    if !strings.Contains(contentStr, content) {
        t.Errorf("File content missing body")
    }
}

func TestIncrementalNumbering(t *testing.T) {
    tempDir := t.TempDir()
    dateDir := filepath.Join(tempDir, "2024-01-15")
    os.MkdirAll(dateDir, 0755)
    
    // Create existing files
    os.WriteFile(filepath.Join(dateDir, "001-first.md"), []byte("test"), 0644)
    os.WriteFile(filepath.Join(dateDir, "002-second.md"), []byte("test"), 0644)
    
    nextNum, err := getNextIncrementalNumber(dateDir)
    if err != nil {
        t.Fatalf("getNextIncrementalNumber failed: %v", err)
    }
    
    if nextNum != 3 {
        t.Errorf("Expected next number 3, got %d", nextNum)
    }
}
```

### Integration Testing

```go
// integration_test.go
func TestEndToEndWorkflow(t *testing.T) {
    // Test complete workflow: create → search → append → export
    tempVault := t.TempDir()
    
    // 1. Create note
    config := Config{VaultPath: tempVault, Title: "Integration Test"}
    err := CreateNewNote(config, "Initial content")
    require.NoError(t, err)
    
    // 2. Verify search finds it
    notes := searchNotes(tempVault)
    require.Len(t, notes, 1)
    require.Contains(t, notes[0].Title, "Integration Test")
    
    // 3. Append content  
    config.AppendMode = true
    err = AppendToNote(config, "Additional content")
    require.NoError(t, err)
    
    // 4. Export and verify
    exportConfig := ExportConfig{VaultPath: tempVault}
    err = ExportNotes(exportConfig)
    require.NoError(t, err)
}
```

### Running Tests

```bash
# Run all tests
go test ./pkg/...

# Run with coverage
go test -cover ./pkg/...

# Run specific test
go test -run TestCreateNewNote ./pkg/note

# Run with race detection
go test -race ./pkg/...

# Verbose output
go test -v ./pkg/...
```

## 🎨 Code Style and Conventions

### Go Conventions

**Naming**:
- Exported functions: `CreateNewNote`, `GetContentFromUser`
- Unexported functions: `sanitizeFilename`, `getNextIncrementalNumber` 
- Constants: `DefaultNoteType`, `MaxPreviewLength`
- Interfaces: `ContentProvider`, `NoteStorage`

**Error Handling**:
```go
// Good - context and wrapping
func processFile(path string) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return errors.Wrap(err, fmt.Sprintf("failed to read file %s", path))
    }
    // ...
}

// Avoid - bare errors
func processFile(path string) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return err  // Lost context!
    }
    // ...
}
```

**Logging**:
```go
// Structured logging with context
log.Debug().
    Str("file", filePath).
    Int("size", len(content)).
    Msg("Processing attached file")

// Include relevant fields for debugging
log.Info().
    Str("vault", config.VaultPath).
    Str("title", noteTitle).
    Str("date", targetDate.Format("2006-01-02")).
    Msg("Creating new note")
```

### Configuration Patterns

```go
// Config structs for each package
type Config struct {
    // Required fields first
    VaultPath string
    Title     string
    
    // Optional fields with sensible defaults  
    DateStr      string // defaults to today
    NoteType     string // defaults to "research"
    AppendMode   bool   // defaults to false
    WithMetadata bool   // defaults to false
}

// Constructor functions when needed
func NewConfig(vaultPath, title string) Config {
    return Config{
        VaultPath: vaultPath,
        Title:     title,
        DateStr:   time.Now().Format("2006-01-02"),
        NoteType:  "research",
    }
}
```

## 🔨 Common Tasks

### Adding a New Command Flag

```go
// 1. Add variable in main.go
var newFeatureFlag bool

// 2. Add flag definition in main()
rootCmd.Flags().BoolVar(&newFeatureFlag, "new-feature", false, "Enable new feature")

// 3. Pass to relevant config
noteConfig := note.Config{
    // ... existing fields ...
    NewFeature: newFeatureFlag,
}

// 4. Handle in relevant package
func CreateNewNote(config Config, content string) error {
    if config.NewFeature {
        // Handle new feature logic
    }
    // ...
}
```

### Adding a New Note Type

```go
// 1. No code changes needed! The tool is already generic
# Create notes of any type
add-research --type "meeting" --title "Team Standup"
add-research --type "journal" --title "Daily Reflection"  
add-research --type "project" --title "Architecture Planning"

// 2. To add validation (optional):
func validateNoteType(noteType string) error {
    validTypes := []string{"research", "ideas", "notes", "meeting", "journal"}
    for _, valid := range validTypes {
        if noteType == valid {
            return nil
        }
    }
    return fmt.Errorf("invalid note type: %s", noteType)
}
```

### Adding Search Filters

```go
// In pkg/note/note.go
type SearchConfig struct {
    VaultPath   string
    Query       string    // Text search
    DateFrom    string    // Date range start
    DateTo      string    // Date range end  
    NoteType    string    // Filter by type
    Tags        []string  // Filter by tags
    MinWords    int       // Minimum word count
    MaxWords    int       // Maximum word count
}

func SearchNotesAdvanced(config SearchConfig) ([]NoteInfo, error) {
    var matchingNotes []NoteInfo
    
    err := filepath.WalkDir(config.VaultPath, func(path string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
            return err
        }
        
        noteInfo := extractNoteInfo(path, d)
        
        // Apply filters
        if config.Query != "" && !strings.Contains(strings.ToLower(noteInfo.Title), strings.ToLower(config.Query)) {
            return nil  // Skip this note
        }
        
        if config.DateFrom != "" {
            fromDate, _ := time.Parse("2006-01-02", config.DateFrom)
            noteDate, _ := time.Parse("2006-01-02", noteInfo.Date)
            if noteDate.Before(fromDate) {
                return nil
            }
        }
        
        // ... other filters ...
        
        matchingNotes = append(matchingNotes, noteInfo)
        return nil
    })
    
    return matchingNotes, err
}
```

### Debugging Tips

**1. Enable Debug Logging**:
```bash
add-research --log-level debug --title "Test" --message "Debug test"
```

**2. Add Temporary Debug Statements**:
```go
log.Debug().
    Interface("config", config).
    Str("content", content[:min(50, len(content))]).
    Msg("Debug: function entry point")
```

**3. Test with Temporary Vault**:
```bash
# Use a test directory
mkdir /tmp/test-vault
export VAULT_PATH=/tmp/test-vault
add-research --title "Test Note" --message "Testing in isolation"
```

## 🐛 Troubleshooting

### Common Issues

**Issue: "Failed to create date directory"**
```bash
# Check permissions
ls -la ~/code/wesen/obsidian-vault/
chmod 755 ~/code/wesen/obsidian-vault/

# Or use different path
add-research --vault-path ~/Documents/notes
```

**Issue: "No notes found" in search**
```bash
# Verify vault structure
find ~/code/wesen/obsidian-vault -name "*.md" -type f

# Check note type directory
ls ~/code/wesen/obsidian-vault/research/
```

**Issue: Clipboard not working**
```bash
# Linux: Install xclip or xsel
sudo apt install xclip

# macOS: Should work out of box
# Windows: Should work out of box

# Test clipboard
echo "test" | add-research --clip --title "Clipboard Test"
```

**Issue: Interactive prompts not working**
```bash
# Check if running in proper terminal
echo $TERM

# Try forcing terminal mode
TERM=xterm-256color add-research

# Debug TUI issues
add-research --log-level debug 2>/tmp/debug.log
```

### Development Issues

**Issue: "Cannot find package"**
```bash
# Update dependencies
go mod tidy
go mod download

# Verify module path
grep module go.mod
```

**Issue: Tests failing**
```bash
# Run specific test with verbose output
go test -v -run TestCreateNewNote ./pkg/note

# Check for race conditions
go test -race ./pkg/...

# Clean test cache
go clean -testcache
```

**Issue: Build errors**
```bash
# Check Go version
go version  # Should be 1.23+

# Verify all imports
go mod why github.com/charmbracelet/huh

# Clean and rebuild
go clean
go build ./...
```

---

## 🎯 Next Steps

After reading this guide, you should be able to:

1. **Understand** the overall architecture and data flow
2. **Navigate** the codebase confidently  
3. **Add** new features following established patterns
4. **Test** your changes thoroughly
5. **Debug** issues effectively

### Recommended Learning Path

1. **Start Small**: Add a simple flag or modify existing behavior
2. **Read Tests**: Understand expected behavior from test cases
3. **Add Features**: Implement a new search filter or content source
4. **Contribute**: Submit a PR with proper tests and documentation

### Resources

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Huh TUI Components](https://github.com/charmbracelet/huh)
- [Zerolog Logging](https://github.com/rs/zerolog)
- [Go Testing Best Practices](https://go.dev/blog/testing)

**Questions?** Open an issue or join the discussion!
