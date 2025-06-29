# Add Research Tool Demos

This directory contains VHS recordings demonstrating the key features of the add-research tool.

## Demo Files

### 1. Basic Note Creation (`demo-basic.*`)
- **File**: `demo-basic.tape` → `demo-basic.gif`
- **Shows**: Interactive title input, content creation, automatic file organization
- **Key Features**: Simple note workflow, structured content, date-based organization

### 2. File Attachments (`demo-files.*`)
- **File**: `demo-files.tape` → `demo-files.gif`  
- **Shows**: Attaching files with `-f` flag, file browser with `-b` flag
- **Key Features**: Multiple file attachment, syntax highlighting, embedded content

### 3. Content Sources (`demo-content.*`)
- **File**: `demo-content.tape` → `demo-content.gif`
- **Shows**: Clipboard integration (`--clip`), piped input, combined sources
- **Key Features**: Multiple input methods, content combination strategies

### 4. Search & Export (`demo-search.*`)
- **File**: `demo-search.tape` → `demo-search.gif`
- **Shows**: Search functionality (`--search`), export with date filtering
- **Key Features**: Fuzzy search, clipboard integration, filtered exports

### 5. Types & Organization (`demo-types.*`)
- **File**: `demo-types.tape` → `demo-types.gif`
- **Shows**: Different note types, custom dates, vault organization, metadata
- **Key Features**: Multi-type support, append mode, structured organization

## Usage

### Running the Demos

```bash
# Generate all GIFs from tape files
vhs demos/demo-basic.tape
vhs demos/demo-files.tape
vhs demos/demo-content.tape
vhs demos/demo-search.tape
vhs demos/demo-types.tape
```

### Viewing Results

- **GIF files**: Ready for web embedding and documentation
- **TXT files**: ASCII screenshots for validation and debugging
- Each demo is ~30 seconds, optimized for web viewing

### Prerequisites

- VHS by Charm Bracelet installed
- add-research tool built (`go build -o add-research .`)
- Proper vault directory structure

## Demo Highlights

### Technical Features Showcased
- ✅ Interactive TUI components
- ✅ File attachment with syntax highlighting  
- ✅ Multiple content input methods
- ✅ Search with fuzzy matching
- ✅ Export with date filtering
- ✅ Metadata and YAML frontmatter
- ✅ Multi-type note organization
- ✅ Append mode functionality

### User Experience Highlights
- 🎯 Quick note creation workflow
- 📁 Automatic file organization
- 🔍 Efficient search and retrieval
- 📊 Rich metadata integration
- 🔄 Flexible content sources
- 📤 Export capabilities

## Integration

These demos can be embedded in:
- GitHub README files
- Documentation websites
- Blog posts and articles
- Product showcases
- Training materials

## File Sizes

| Demo | GIF Size | Duration | Key Features |
|------|----------|----------|-------------|
| Basic | ~696KB | ~30s | Core workflow |
| Files | ~4.2MB | ~30s | Attachments |
| Content | ~7.6MB | ~30s | Input methods |
| Search | ~8.4MB | ~30s | Search/Export |
| Types | ~13MB | ~30s | Organization |

Total: ~33MB for complete demo suite

## Recent Improvements (June 2025)

### Fixed VHS Demo Issues
- ✅ **Timing optimization**: Proper typing speeds for different content types
- ✅ **Content handling**: Simplified multiline content to avoid VHS parsing errors
- ✅ **Command reliability**: All commands now work correctly in non-TTY environments
- ✅ **Interactive features**: Replaced problematic interactive prompts with command-line flags
- ✅ **File validation**: All demos create and display actual working examples

### Technical Improvements
- **Escape handling**: Fixed quote and multiline content escaping in VHS scripts
- **Error resilience**: Commands designed to work even if previous steps fail
- **Content verification**: Each demo shows actual file contents and outputs
- **Cleanup procedures**: Proper cleanup of temporary files after demos
- **Performance**: Optimized sleep times and command execution flow
