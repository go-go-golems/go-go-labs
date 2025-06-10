# Playbook Manager Design Document

## Overview

Playbook Manager is a CLI/TUI tool for managing "playbooks" - contextual documents that can be passed as additional context to LLMs to help with work. The tool maintains a central registry of playbooks and collections, allowing users to organize, search, and deploy these documents into workspaces as needed.

## Core Concepts

### Entities
The system treats both playbooks and collections as first-class entities with identical metadata capabilities:

- **Playbooks**: Individual documents with content (markdown files, guides, standards, etc.)
- **Collections**: Organized groups of playbooks and/or other collections, enabling hierarchical organization

### Registry
All data is stored in a single SQLite database (`~/.playbooks/registry.db`) making the system self-contained and portable.

## Database Schema

```sql
-- Unified entities table (both playbooks and collections)
CREATE TABLE entities (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('playbook', 'collection')),
    title TEXT NOT NULL,
    description TEXT,
    summary TEXT,
    canonical_url TEXT,              -- NULL for collections, required for playbooks
    content TEXT,                    -- NULL for collections, required for playbooks
    content_hash TEXT,               -- NULL for collections
    filename TEXT,                   -- NULL for collections
    tags TEXT,                       -- JSON array of tags
    last_fetched DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE entity_metadata (
    entity_id INTEGER REFERENCES entities(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (entity_id, key)
);

-- Collection membership (what playbooks/collections are in a collection)
CREATE TABLE collection_members (
    collection_id INTEGER REFERENCES entities(id) ON DELETE CASCADE,
    member_id INTEGER REFERENCES entities(id) ON DELETE CASCADE,
    relative_path TEXT,              -- Optional: organize within collection
    PRIMARY KEY (collection_id, member_id),
    CHECK ((SELECT type FROM entities WHERE id = collection_id) = 'collection')
);

CREATE TABLE deployments (
    id INTEGER PRIMARY KEY,
    entity_id INTEGER REFERENCES entities(id),
    target_directory TEXT,
    deployed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Command Line Interface

### Registration and Creation

```bash
# Register a playbook from file or URL
pb register <path|url> \
  --title "Title" \
  --description "Detailed description" \
  --summary "Brief summary" \
  --meta key=value \
  --meta key2=value2 \
  --tags tag1,tag2 \
  --filename override.md

# Create a new collection
pb create collection "Collection Name" \
  --title "Display Title" \
  --description "Purpose and contents" \
  --summary "Brief overview" \
  --meta key=value \
  --tags tag1,tag2
```

### Discovery and Search

```bash
# List entities with optional filters
pb list [--type playbook|collection] [--tags tag1,tag2] [--meta key=value] [--stale]

# Search across titles, descriptions, summaries, and content
pb search "query" [--type playbook|collection] [--meta-key key] [--in-title] [--in-summary]

# Show detailed information about an entity
pb show <id>
```

### Collection Management

```bash
# Add members to collections
pb add <collection-id> <playbook-id> [--path subdir/file.md]
pb add <collection-id> <other-collection-id>

# Remove members from collections
pb remove <collection-id> <member-id>
```

### Metadata Management

```bash
# Set metadata on any entity
pb meta set <id> <key> <value>

# Get metadata from any entity
pb meta get <id> [key]

# Remove metadata from any entity
pb meta remove <id> <key>
```

### Content Management

```bash
# Update playbooks from their canonical sources
pb update <playbook-id|--all>

# Remove entities
pb remove <id>
```

### Deployment

```bash
# Deploy individual playbooks or entire collections
pb deploy <id> <target-directory> [--filename override.md]

# Check deployment status
pb status [target-directory]
```

## Key Features

### Unified Entity Model
Both playbooks and collections use the same metadata system, making discovery and organization consistent across entity types.

### Rich Metadata System
- **Core fields**: title, description, summary
- **Flexible key-value metadata**: repository info, authorship, versioning, categorization
- **Auto-populated metadata**: source type, file size, fetch dates, git information

### Content Integrity
- SHA256 hashing of playbook content
- Change detection for source updates
- Last fetched timestamps for staleness tracking

### Hierarchical Organization
- Collections can contain playbooks and other collections
- Optional relative paths for organizing files within collections
- Flexible tagging system for cross-cutting categorization

### Self-Contained Storage
- All content stored directly in SQLite database
- No external file dependencies
- Easy backup and synchronization

## Example Workflows

### Basic Playbook Management
```bash
# Register a coding standards document
pb register ./docs/coding-standards.md \
  --title "Python Coding Standards" \
  --summary "Style guide and best practices for Python development" \
  --meta repository=github.com/myorg/standards \
  --meta version=2.1 \
  --tags python,standards

# Deploy to workspace
pb deploy 1 ./my-project/playbooks/
```

### Collection Organization
```bash
# Create a backend development collection
pb create collection "Backend Development" \
  --summary "All backend development guidelines and standards" \
  --meta team=backend \
  --tags backend,standards

# Add playbooks to the collection
pb add 2 1  # Add coding standards to backend collection
pb add 2 3  # Add API guidelines
pb add 2 4  # Add database standards

# Deploy entire collection
pb deploy 2 ./new-project/playbooks/
```

### Discovery and Maintenance
```bash
# Find all Python-related content
pb list --tags python

# Find stale playbooks that might need updates
pb list --stale

# Update all playbooks from their sources
pb update --all
```
