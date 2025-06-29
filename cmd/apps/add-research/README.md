# 📝 Add Research Tool

> A powerful CLI tool for creating, organizing, and managing research notes in Obsidian-compatible markdown format

![Demo placeholder - add gif here](./demo.gif)

## 🎯 What It Does

**Add Research** is an interactive command-line tool that helps you quickly capture, organize, and retrieve research notes. It automatically structures your notes by date, supports file attachments, handles links, and integrates seamlessly with your Obsidian vault workflow.

Perfect for researchers, developers, students, and anyone who needs to quickly capture and organize information with proper metadata and searchable content.

## ✨ Key Features

- 📅 **Smart Organization** - Automatically organizes notes by date in `YYYY-MM-DD/NNN-title.md` format
- 🔍 **Intelligent Search** - Fuzzy search with content previews and metadata
- 📎 **File Attachments** - Attach and embed multiple files with syntax highlighting
- 🔗 **Link Management** - Interactive or batch link collection with smart URL handling
- 📊 **Metadata Rich** - Auto-generated YAML frontmatter with tags, timestamps, and word counts
- 📤 **Export System** - Combine notes into single markdown files with date filtering
- 🖥️ **Multiple Input Methods** - Command line, interactive prompts, clipboard, or stdin
- 🗂️ **File Browser** - Interactive file selection with tree navigation
- ⚙️ **Configurable** - YAML configuration file support
- 🏷️ **Note Types** - Support for research, ideas, notes, and custom types

## 🚀 Installation

### Prerequisites

- Go 1.23+ installed
- Access to the go-go-labs repository

### Build from Source

```bash
# Clone the repository
git clone https://github.com/go-go-golems/go-go-labs.git
cd go-go-labs

# Build the tool
go build -o add-research ./cmd/apps/add-research

# Move to your PATH (optional)
sudo mv add-research /usr/local/bin/
```

### Quick Test

```bash
# Verify installation
add-research --help
```

## 🏃 Quick Start

### Create Your First Note

```bash
# Interactive mode (default)
add-research

# Quick note with title
add-research --title "API Research" --message "Found new REST API patterns"

# From clipboard
add-research --clip --title "Clipboard Content"
```

### Basic Workflow

1. **Run the tool**: `add-research`
2. **Enter title**: When prompted, type your research note title
3. **Add content**: Type or paste your content (Ctrl+D to finish)
4. **Add links**: Enter relevant URLs (optional, press Enter to skip)
5. **Done!** Your note is saved with automatic organization

### Common Commands

```bash
# Search existing notes
add-research --search

# Append to today's notes
add-research --append

# Create with metadata
add-research --metadata --title "Important Research"

# Export date range
add-research --export --export-from "2024-01-01" --export-to "2024-12-31"
```

## 🎯 Feature Showcase

### 📅 Basic Note Creation

Create structured research notes with automatic organization:

```bash
# Simple note creation
add-research --title "GraphQL Best Practices" --message "
## Key Findings
- Use fragments for reusable queries
- Implement proper error handling
- Cache query results effectively
"
```

**Result**: Creates `~/code/wesen/obsidian-vault/research/2024-01-15/001-GraphQL-Best-Practices.md`

### 📎 File Attachments

Attach and embed files directly into your notes:

```bash
# Attach specific files
add-research --file "config.yaml" --file "schema.graphql" --title "API Configuration"

# Interactive file browser
add-research --browse --title "Project Files Review"
```

**What happens**: Files are embedded with proper syntax highlighting and metadata.

### 🔍 Search and Clipboard Integration

Find and reuse your research efficiently:

```bash
# Search with preview
add-research --search
# Shows: "2024-01-15 - GraphQL Best Practices (245 words, 1.2KB)
#          Use fragments for reusable queries, implement proper..."

# Copy found content to clipboard automatically
# Select note → Choose "Copy to clipboard?" → Yes
```

### 🏷️ Different Note Types

Organize by category:

```bash
# Ideas notebook
add-research --type "ideas" --title "Mobile App Concept"

# Meeting notes
add-research --type "meetings" --title "Team Sync Jan 15"

# Technical notes
add-research --type "technical" --title "Database Migration Strategy"
```

### 🔗 Link Management

Three flexible approaches to handle relevant links:

```bash
# 1. Interactive (default) - asks for links
add-research --title "Research Topic"

# 2. Provide links directly  
add-research --links "https://api.github.com" "https://docs.graphql.org" --title "API References"

# 3. Skip links entirely
add-research --no-links --title "Quick Note"
```

### 📊 Metadata and YAML Frontmatter

Rich metadata for better organization:

```bash
add-research --metadata --title "Important Research"
```

**Generates**:
```yaml
---
title: "Important Research"
id: "important-research-a1b2c3d4"
slug: "important-research-a1b2c3d4"
date: 2024-01-15
type: research
tags:
  - type/research
  - year/2024
  - month/01
created: 2024-01-15T10:30:00Z
modified: 2024-01-15T10:30:00Z
source: "add-research-tool"
word_count: 0
---
```

### 📤 Export Functionality

Combine and export your research:

```bash
# Export all notes
add-research --export

# Export date range
add-research --export --export-from "2024-01-01" --export-to "2024-01-31" --export-path "january-research.md"

# Export specific type
cd ~/code/wesen/obsidian-vault/ideas
add-research --export --export-path "all-ideas.md"
```

## ⚙️ Configuration

### Configuration File

Create `~/.add-research.yaml`:

```yaml
vault_base_path: "~/Documents/obsidian-vault"
default_note_type: "research"
with_metadata: true
ask_for_links: true
```

### Command-Line Options

```bash
# Core Options
--title, -t         Set note title (skips interactive input)
--message, -m       Provide content directly
--date              Use specific date (YYYY-MM-DD, default: today)
--type              Note type (research, ideas, notes, etc.)

# Content Options
--clip, -c          Use clipboard content
--file, -f          Attach files (multiple allowed)
--browse, -b        Interactive file browser

# Link Options
--links             Provide links directly (skips prompting)
--no-links          Disable link prompting entirely
--ask-links         Prompt for links (deprecated - now default)

# Modes
--append            Append to existing note
--search, -s        Search existing notes
--export            Export notes to combined file

# Export Options
--export-path       Output file path
--export-from       Start date for export (YYYY-MM-DD)
--export-to         End date for export (YYYY-MM-DD)

# Other
--metadata          Include YAML frontmatter
--config            Config file path
--log-level         Logging level (debug, info, warn, error)
```

## 💡 Use Cases

### 📚 Research Workflow

**Daily Research Collection**:
```bash
# Morning: Capture articles from clipboard
add-research --clip --title "Morning Reading"

# Afternoon: Add findings with attachments
add-research --browse --title "Experiment Results"

# Evening: Review and append to existing notes
add-research --search  # Find relevant note
add-research --append --date "2024-01-15"
```

### 🔬 Academic Research

**Literature Review Process**:
```bash
# Capture paper summaries with links
add-research --type "papers" --links "https://arxiv.org/abs/2401.12345" --title "Neural Network Architecture Study"

# Export monthly reviews
add-research --export --export-from "2024-01-01" --export-to "2024-01-31" --export-path "january-papers.md"
```

### 💻 Software Development

**Technical Documentation**:
```bash
# Document API discoveries
add-research --type "technical" --file "api-response.json" --title "New API Endpoints"

# Meeting notes with action items
add-research --type "meetings" --metadata --title "Architecture Review"
```

### 🎓 Learning and Education

**Study Session Notes**:
```bash
# Quick concept capture
add-research --type "study" --title "Design Patterns" --message "
## Observer Pattern
- Used for event handling
- Loose coupling between objects
"

# Attach code examples
add-research --file "observer-example.py" --title "Observer Implementation"
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/your-username/go-go-labs.git
cd go-go-labs/cmd/apps/add-research

# Run in development mode
go run main.go --help

# Run tests
go test ./pkg/...
```

### Areas for Contribution

- 🔍 **Enhanced Search**: Vector embeddings, semantic search
- 🎨 **UI Improvements**: Better TUI components, themes
- 📊 **Analytics**: Usage statistics, content insights  
- 🔧 **Integrations**: Other note-taking tools, cloud sync
- 📱 **Mobile**: Terminal-friendly mobile interfaces

### Code Guidelines

- Follow Go conventions and gofmt formatting
- Add tests for new functionality
- Update documentation for new features
- Use structured logging with zerolog
- Handle errors with github.com/pkg/errors wrapping

### Submitting Changes

1. Create a feature branch: `git checkout -b feature/amazing-feature`
2. Make your changes and test thoroughly
3. Run the formatter: `go fmt ./...`
4. Add tests: `go test ./...`
5. Commit with clear messages
6. Submit a pull request

---

**📧 Questions?** Open an issue or start a discussion!

**⭐ Like this tool?** Give us a star on GitHub!
