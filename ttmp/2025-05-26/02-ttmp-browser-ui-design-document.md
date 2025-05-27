---
title: "TTMP File Management System - Design Document"
date: "2025-05-26"
author: "manuel"
tags: ["design", "ttmp", "go", "web-ui", "file-management"]
status: "draft"
---

# TTMP File Management System - Design Document

## Overview

A web-based file management system for organizing and browsing temporary markdown files stored in the `ttmp/` directory structure. The system will provide search, tagging, creation, and metadata management capabilities through a clean web interface.

## Current State Analysis

### Existing Structure
- Files organized by date: `ttmp/YYYY-MM-DD/NN-description.md`
- Primarily markdown files containing research, tutorials, and documentation
- No consistent metadata structure currently
- Manual file organization and discovery

### Prototype Analysis
The existing React prototype (`01-library-prototype-from-claude.tsx`) demonstrates:
- Sidebar with folders, search, and tag filtering
- Chat list with preview and metadata
- Detail view with notes and full content
- Favorite/star functionality
- Tag-based organization

## Requirements

### Functional Requirements

1. **File Discovery & Browsing**
   - List all ttmp files organized by date
   - Search by filename, content, and metadata
   - Filter by tags, date ranges, and folders
   - Sort by date, title, or relevance

2. **Metadata Management**
   - YAML frontmatter for file metadata
   - Support for title, tags, author, date, status fields
   - Editable metadata through web interface
   - Automatic metadata extraction and indexing

3. **File Operations**
   - Create new files with templates
   - Edit existing files (metadata and content)
   - Delete files with confirmation
   - Rename/move files within date structure

4. **Organization Features**
   - Tag-based categorization
   - Favorite/star files
   - Notes/annotations on files
   - Date-based folder navigation

5. **Search & Filtering**
   - Full-text search across content
   - Metadata-based filtering
   - Tag intersection/union filtering
   - Date range filtering

### Non-Functional Requirements

1. **Performance**
   - Fast file indexing and search
   - Responsive UI for large file collections
   - Efficient file watching for changes

2. **Usability**
   - Clean, intuitive interface
   - Keyboard shortcuts for common operations
   - Mobile-responsive design

3. **Reliability**
   - Safe file operations with backups
   - Graceful handling of malformed files
   - Atomic metadata updates

## Technical Architecture

### Technology Stack

- **Backend**: Go with Cobra CLI framework
- **Web Framework**: Standard library HTTP server with custom routing
- **Templates**: Templ for type-safe HTML generation
- **Frontend**: Bootstrap CSS + vanilla JavaScript
- **File Format**: Markdown with YAML frontmatter
- **Search**: In-memory indexing with optional file-based persistence

### Project Structure

```
ttmp-browser/
├── cmd/
│   └── root.go                 # Cobra CLI setup
├── pkg/
│   ├── models/
│   │   ├── file.go            # File metadata model
│   │   └── index.go           # Search index model
│   ├── services/
│   │   ├── fileservice.go     # File operations
│   │   ├── indexservice.go    # Search and indexing
│   │   └── watchservice.go    # File system watching
│   ├── handlers/
│   │   ├── api.go             # REST API handlers
│   │   ├── pages.go           # Page handlers
│   │   └── middleware.go      # Common middleware
│   └── templates/
│       ├── layout.templ       # Base layout
│       ├── index.templ        # Main browser page
│       ├── file.templ         # File detail view
│       └── components/        # Reusable components
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   ├── bootstrap.min.css
│   │   │   └── app.css        # Custom styles
│   │   └── js/
│   │       ├── bootstrap.min.js
│   │       └── app.js         # Application JavaScript
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Data Models

#### File Metadata
```go
type FileMetadata struct {
    Path         string    `yaml:"-" json:"path"`
    Title        string    `yaml:"title" json:"title"`
    Date         time.Time `yaml:"date" json:"date"`
    Author       string    `yaml:"author" json:"author"`
    Tags         []string  `yaml:"tags" json:"tags"`
    Status       string    `yaml:"status" json:"status"`
    Description  string    `yaml:"description" json:"description"`
    Favorite     bool      `yaml:"favorite" json:"favorite"`
    Notes        string    `yaml:"notes" json:"notes"`
    
    // Computed fields
    ModTime      time.Time `yaml:"-" json:"modTime"`
    Size         int64     `yaml:"-" json:"size"`
    ContentHash  string    `yaml:"-" json:"contentHash"`
    Preview      string    `yaml:"-" json:"preview"`
}
```

#### Search Index
```go
type SearchIndex struct {
    Files       map[string]*FileMetadata
    TagIndex    map[string][]string      // tag -> file paths
    ContentIndex map[string][]string     // word -> file paths
    DateIndex   map[string][]string      // date -> file paths
    mutex       sync.RWMutex
}
```

### API Design

#### REST Endpoints

```
GET    /                           # Main browser page
GET    /api/files                  # List files with filtering
GET    /api/files/{path}           # Get specific file
PUT    /api/files/{path}           # Update file metadata/content
POST   /api/files                  # Create new file
DELETE /api/files/{path}           # Delete file
GET    /api/search                 # Search files
GET    /api/tags                   # Get all tags
POST   /api/files/{path}/favorite  # Toggle favorite
```

#### Query Parameters for `/api/files`
- `search`: Full-text search query
- `tags`: Comma-separated tag list
- `date_from`, `date_to`: Date range filtering
- `favorites`: Show only favorites
- `sort`: Sort order (date, title, relevance)
- `limit`, `offset`: Pagination

### Frontend Architecture

#### JavaScript Modules
```javascript
// app.js - Main application logic
class TTMPBrowser {
    constructor() {
        this.searchIndex = new SearchManager();
        this.fileManager = new FileManager();
        this.ui = new UIManager();
    }
}

class SearchManager {
    // Handle search, filtering, and sorting
}

class FileManager {
    // Handle file operations and API calls
}

class UIManager {
    // Handle DOM manipulation and events
}
```

#### CSS Organization
```css
/* app.css */
:root {
    /* Custom CSS variables for theming */
}

/* Layout components */
.ttmp-sidebar { }
.ttmp-main { }
.ttmp-detail { }

/* File components */
.file-card { }
.file-metadata { }
.tag-list { }

/* Utility classes */
.text-truncate { }
.fade-in { }
```

## Implementation Plan

### Phase 1: Core Backend (Week 1)
1. **Project Setup**
   - [ ] Initialize Go module with Cobra
   - [ ] Set up basic project structure
   - [ ] Configure Templ and build system

2. **File Service**
   - [ ] Implement file discovery and parsing
   - [ ] YAML frontmatter parsing
   - [ ] Basic CRUD operations

3. **Search Index**
   - [ ] In-memory indexing system
   - [ ] Full-text search implementation
   - [ ] Tag and metadata indexing

### Phase 2: Web Interface (Week 2)
1. **Templates & Styling**
   - [ ] Base layout with Bootstrap
   - [ ] Main browser page template
   - [ ] File detail view template
   - [ ] Custom CSS for ttmp-specific styling

2. **API Handlers**
   - [ ] REST API implementation
   - [ ] File listing and filtering
   - [ ] Search endpoint
   - [ ] File operations (CRUD)

### Phase 3: Frontend Interactivity (Week 3)
1. **JavaScript Application**
   - [ ] Search and filtering UI
   - [ ] File operations (create, edit, delete)
   - [ ] Tag management
   - [ ] Favorites functionality

2. **Advanced Features**
   - [ ] File watching for live updates
   - [ ] Keyboard shortcuts
   - [ ] Responsive design improvements

### Phase 4: Polish & Optimization (Week 4)
1. **Performance**
   - [ ] Search optimization
   - [ ] Lazy loading for large collections
   - [ ] Caching strategies

2. **User Experience**
   - [ ] Error handling and validation
   - [ ] Loading states and feedback
   - [ ] Help documentation

## File Format Specification

### YAML Frontmatter Schema
```yaml
---
title: "Human-readable title"
date: "2025-05-26"
author: "manuel"
tags: ["tag1", "tag2", "tag3"]
status: "draft|in-progress|complete|archived"
description: "Brief description of the content"
favorite: false
notes: "Personal notes about this file"
---

# Markdown Content

The actual content of the file follows the frontmatter.
```

### Default Templates
- **Research Note**: Basic template with research-focused metadata
- **Tutorial**: Template for step-by-step guides
- **Meeting Notes**: Template for meeting documentation
- **Project Spec**: Template for project specifications
