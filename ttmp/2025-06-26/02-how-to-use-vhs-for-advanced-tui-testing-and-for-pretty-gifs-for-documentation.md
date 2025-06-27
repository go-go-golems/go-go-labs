# The Complete Guide to VHS for Advanced TUI Testing and Documentation

**Author**: AI Assistant  
**Date**: June 26, 2025  
**Project**: Advanced TUI Development with Bubbletea File Picker

## Table of Contents

1. [Introduction to VHS](#introduction-to-vhs)
2. [Installation and Setup](#installation-and-setup)
3. [Basic VHS Syntax](#basic-vhs-syntax)
4. [Advanced Testing Techniques](#advanced-testing-techniques)
5. [Validation Workflows](#validation-workflows)
6. [Creating Documentation GIFs](#creating-documentation-gifs)
7. [Real-World Examples](#real-world-examples)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)
10. [Integration with Development Workflow](#integration-with-development-workflow)

---

## Introduction to VHS

**VHS (Video Hypterterminal Simulator)** by Charm is a powerful tool for creating terminal recordings and automating TUI testing. It's particularly valuable for:

- **Automated Testing**: Validate TUI functionality without manual intervention
- **Documentation**: Create beautiful GIFs for README files and documentation
- **Regression Testing**: Ensure UI changes don't break existing functionality
- **Feature Validation**: Test complex user interactions programmatically
- **Visual Debugging**: Capture screenshots at specific moments for debugging

### Why VHS for TUI Development?

1. **Deterministic**: Same script produces identical results every time
2. **Fast**: Much faster than manual testing
3. **Visual**: Creates both text output and visual recordings
4. **Scriptable**: Easily integrate into CI/CD pipelines
5. **Screenshot Capable**: Take precise snapshots for validation

---

## Installation and Setup

### Install VHS

```bash
# macOS with Homebrew
brew install vhs

# Or with Go
go install github.com/charmbracelet/vhs@latest

# Verify installation
vhs --version
```

### Install Dependencies

VHS requires a few additional tools:

```bash
# Install ttyd (for web-based terminal simulation)
brew install ttyd

# Install ImageMagick (for image processing)
brew install imagemagick

# Install a font that supports Unicode (recommended)
brew tap homebrew/cask-fonts
brew install font-fira-code-nerd-font
```

### Project Structure for VHS

```
your-project/
├── demo/                    # VHS scripts and outputs
│   ├── scripts/            # .tape files
│   ├── screenshots/        # .txt outputs for validation
│   ├── gifs/              # .gif outputs for documentation
│   ├── run-all-demos.sh   # Batch script runner
│   └── setup-test-env.sh  # Test environment setup
├── your-app              # Your TUI application
└── test-files/           # Test data structure
```

---

## Basic VHS Syntax

### Core VHS Commands

```bash
# VHS Tape Script (.tape file)

# Output configuration
Output demo.gif              # Create GIF
Output demo.txt              # Create text output
Output demo.mp4              # Create MP4 video

# Terminal configuration
Set FontSize 14              # Font size
Set Width 1200               # Terminal width in pixels
Set Height 800               # Terminal height in pixels
Set Theme "Dracula"          # Color theme

# Typing and input
Type "ls -la"                # Type text
Enter                        # Press Enter key
Space                        # Press Space key
Backspace                    # Press Backspace
Tab                          # Press Tab key
Escape                       # Press Escape key

# Special keys
Up                           # Arrow up
Down                         # Arrow down
Left                         # Arrow left
Right                        # Arrow right
Home                         # Home key
End                          # End key

# Complex key combinations
Ctrl+C                       # Control combinations
Alt+Left                     # Alt combinations
Shift+Tab                    # Shift combinations

# Timing control
Sleep 2s                     # Wait 2 seconds
Sleep 500ms                  # Wait 500 milliseconds

# Screenshots
Screenshot output.txt        # Take text screenshot
```

### Basic Example

```bash
# demo/basic-example.tape
Output demo/basic-example.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

Type "./my-app"
Enter
Sleep 2s

Type "hello world"
Enter
Sleep 1s

Type "q"
```

---

## Advanced Testing Techniques

### 1. Multi-Step Validation Testing

Create comprehensive test scenarios that validate complex workflows:

```bash
# demo/validate-file-operations.tape
Output demo/validate-file-operations.txt
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"

# Launch application
Type "./filepicker test-files"
Enter
Sleep 2s

# Take initial state screenshot
Screenshot demo/initial-state.txt

# Navigate and select files
Down Down Down
Space                        # Select file
Sleep 300ms
Down
Space                        # Select another file
Sleep 300ms

# Take selection screenshot
Screenshot demo/files-selected.txt

# Copy files
Type "c"
Sleep 500ms

# Navigate to destination
Backspace
Sleep 500ms
Down
Enter
Sleep 1s

# Paste files
Type "v"
Sleep 1s

# Take final state screenshot
Screenshot demo/files-copied.txt

Type "q"
```

### 2. Responsive Layout Testing

Test different window sizes to validate responsive behavior:

```bash
# demo/test-responsive-narrow.tape
Set Width 600                # Test narrow layout
Set Height 400
Set Theme "Dracula"
Output demo/responsive-narrow.txt

Type "./app"
Enter
Sleep 2s
Screenshot demo/narrow-layout.txt
Type "q"
```

```bash
# demo/test-responsive-wide.tape
Set Width 1600               # Test wide layout
Set Height 800
Set Theme "Dracula"
Output demo/responsive-wide.txt

Type "./app"
Enter
Sleep 2s
Screenshot demo/wide-layout.txt
Type "q"
```

### 3. Error Condition Testing

Test error handling and edge cases:

```bash
# demo/test-error-handling.tape
Output demo/error-handling.txt
Set FontSize 12
Set Width 1000
Set Height 600
Set Theme "Dracula"

Type "./app /nonexistent/path"
Enter
Sleep 2s

# Take screenshot of error message
Screenshot demo/error-message.txt

Type "q"
```

### 4. Performance Testing

Test with large datasets:

```bash
# demo/test-large-directory.tape
Output demo/performance-test.txt
Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "Dracula"

# Create large test directory first
Type "mkdir large-test && cd large-test"
Enter
Sleep 500ms

Type "for i in {1..1000}; do touch file$i.txt; done"
Enter
Sleep 3s

Type "../filepicker ."
Enter
Sleep 2s

# Test scrolling performance
Home                         # Jump to top
Sleep 500ms
End                          # Jump to bottom
Sleep 500ms

# Take screenshot showing large directory handling
Screenshot demo/large-directory.txt

Type "q"
```

---

## Validation Workflows

### 1. Screenshot-Based Validation

Use text screenshots to validate UI state:

```bash
#!/bin/bash
# validate-ui-state.sh

# Run VHS script
vhs demo/test-functionality.tape

# Check if expected text appears in screenshots
if grep -q "Expected Text" demo/screenshot.txt; then
    echo "✅ UI state validation passed"
else
    echo "❌ UI state validation failed"
    exit 1
fi
```

### 2. Text Content Validation

Validate specific content in screenshots:

```bash
# demo/validate-file-list.tape
Output demo/validate-file-list.txt
Set FontSize 12
Set Width 1000
Set Height 600
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot for validation
Screenshot demo/file-list.txt

Type "q"
```

Then validate:

```bash
#!/bin/bash
# Check for expected files
expected_files=("config.ini" "docker-compose.yml" "test-script.sh")

for file in "${expected_files[@]}"; do
    if grep -q "$file" demo/file-list.txt; then
        echo "✅ Found $file"
    else
        echo "❌ Missing $file"
        exit 1
    fi
done
```

### 3. State Transition Testing

Test complex state changes:

```bash
# demo/test-state-transitions.tape
Output demo/state-transitions.txt
Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# State 1: Normal view
Screenshot demo/state-normal.txt

# State 2: Search mode
Type "/"
Sleep 500ms
Screenshot demo/state-search.txt

# State 3: Help mode
Escape
Sleep 500ms
Type "?"
Sleep 1s
Screenshot demo/state-help.txt

# State 4: Back to normal
Type "?"
Sleep 500ms
Screenshot demo/state-normal-again.txt

Type "q"
```

---

## Creating Documentation GIFs

### 1. Feature Demonstration GIFs

Create compelling demos for documentation:

```bash
# demo/feature-showcase.tape
Output demo/feature-showcase.gif
Set FontSize 13
Set Width 1200
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Show basic navigation
Down Down Down
Sleep 800ms

# Show multi-selection
Space
Sleep 500ms
Down
Space
Sleep 500ms

# Show file operations
Type "c"
Sleep 800ms

# Navigate and paste
Backspace
Sleep 500ms
Down
Enter
Sleep 500ms
Type "v"
Sleep 1s

# Show result
Down Down
Sleep 1s

Escape
```

### 2. Progressive Feature Demos

Show features building on each other:

```bash
# demo/progressive-features.tape
Output demo/progressive-features.gif
Set FontSize 12
Set Width 1400
Set Height 800
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Step 1: Basic navigation
Down Down
Sleep 1s

# Step 2: Show preview panel
Tab
Sleep 1s

# Step 3: Show search
Type "/"
Sleep 500ms
Type "doc"
Sleep 1s

# Step 4: Show detailed view
Escape
Sleep 500ms
Type "F3"
Sleep 1s

# Step 5: Show sorting
Type "F4"
Sleep 1s
Type "F4"
Sleep 1s

Escape
```

### 3. Quick Feature Highlights

Create short, focused demos:

```bash
# demo/quick-copy-paste.tape
Output demo/quick-copy-paste.gif
Set FontSize 14
Set Width 1000
Set Height 600
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 1s

# Quick demo: select, copy, navigate, paste
Down Down
Space
Sleep 300ms
Type "c"
Sleep 300ms
Enter
Sleep 300ms
Type "v"
Sleep 800ms

Escape
```

---

## Real-World Examples

### Example 1: File Picker Multi-Selection Testing

```bash
# demo/test-multi-selection.tape
Output demo/test-multi-selection.txt
Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Test individual selection
Down Down
Space
Sleep 300ms
Screenshot demo/single-selection.txt

# Test multi-selection
Down
Space
Sleep 300ms
Down
Space
Sleep 300ms
Screenshot demo/multi-selection.txt

# Test select all
Type "a"
Sleep 500ms
Screenshot demo/select-all.txt

# Test deselect all
Type "A"
Sleep 500ms
Screenshot demo/deselect-all.txt

Type "q"
```

### Example 2: Responsive Column Testing

```bash
# demo/test-column-layout.tape
Output demo/test-column-layout.txt
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Show full layout
Screenshot demo/full-columns.txt

Type "q"
```

```bash
# demo/test-narrow-layout.tape
Output demo/test-narrow-layout.txt
Set FontSize 12
Set Width 800
Set Height 600
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Show responsive layout
Screenshot demo/narrow-columns.txt

Type "q"
```

### Example 3: History Navigation Testing

```bash
# demo/test-history-navigation.tape
Output demo/test-history.txt
Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate forward through directories
Down Down
Enter
Sleep 1s
Screenshot demo/step1-documents.txt

Down Down
Enter
Sleep 1s
Screenshot demo/step2-subdirectory.txt

# Test back navigation
Type "h"
Sleep 1s
Screenshot demo/step3-back.txt

Type "h"
Sleep 1s
Screenshot demo/step4-back-to-root.txt

# Test forward navigation
Type "l"
Sleep 1s
Screenshot demo/step5-forward.txt

Type "q"
```

---

## Best Practices

### 1. Script Organization

```bash
# Organize scripts by purpose
demo/
├── testing/               # Validation scripts
│   ├── test-basic-nav.tape
│   ├── test-file-ops.tape
│   └── test-responsive.tape
├── documentation/         # Demo scripts for docs
│   ├── showcase-features.tape
│   ├── quick-demo.tape
│   └── advanced-usage.tape
└── validation/           # Screenshot validation
    ├── validate-ui.tape
    └── validate-content.tape
```

### 2. Timing Best Practices

```bash
# Good timing practices

# Give applications time to start
Type "./app"
Enter
Sleep 2s                  # Always wait for app startup

# Brief pauses for UI updates
Type "command"
Sleep 300ms              # Short pause for immediate feedback

# Longer pauses for complex operations
Type "v"                 # Paste operation
Sleep 1s                 # Wait for file operations

# Screenshot timing
Type "F3"                # Toggle view
Sleep 500ms              # Wait for view to update
Screenshot demo/view.txt # Then take screenshot
```

### 3. Error Handling in Scripts

```bash
# Defensive scripting

# Always end with a way to exit
Type "q"                 # Or Escape, or Ctrl+C

# For interactive apps, provide escape routes
Type "?"                 # Open help
Sleep 2s
Type "?"                 # Close help
Sleep 500ms
Type "q"                 # Exit app
```

### 4. Consistent Formatting

```bash
# Use consistent formatting for readability

# Header comment
# VHS Script: Description of what this tests
# Purpose: Validation/Documentation/Testing

# Configuration block
Output demo/script-name.gif
Set FontSize 12
Set Width 1200
Set Height 700
Set Theme "Dracula"

# Main script with comments
# Step 1: Launch application
Type "./app"
Enter
Sleep 2s

# Step 2: Navigate to feature
Down Down
Enter
Sleep 1s

# Step 3: Test feature
Space
Sleep 300ms

# Cleanup
Type "q"
```

---

## Troubleshooting

### Common Issues and Solutions

#### 1. Application Doesn't Start

```bash
# Problem: App exits immediately
Type "./app"
Enter
Sleep 2s
Screenshot demo/empty.txt    # Results in just shell prompt

# Solution: Check app exists and is executable
Type "ls -la ./app"
Enter
Sleep 1s
Type "./app --help"          # Test with help flag first
Enter
Sleep 2s
```

#### 2. Timing Issues

```bash
# Problem: Screenshots taken before UI updates
Type "F3"
Screenshot demo/immediate.txt  # May not show the change

# Solution: Add appropriate delays
Type "F3"
Sleep 500ms                    # Wait for UI update
Screenshot demo/delayed.txt    # Now captures the change
```

#### 3. Window Size Problems

```bash
# Problem: Content gets cut off
Set Width 800
Set Height 400               # Too small for content

# Solution: Test different sizes
Set Width 1200              # Wider window
Set Height 700               # Taller window

# Or test multiple sizes
Output demo/small.txt
Set Width 600
Set Height 400
# ... test script ...

Output demo/large.txt
Set Width 1600
Set Height 900
# ... same test script ...
```

#### 4. Key Input Issues

```bash
# Problem: Special keys not working
Type "alt+left"              # Might not work

# Solution: Use proper key syntax
Alt+Left                     # Correct syntax
Type "h"                     # Alternative key binding

# For complex combinations
Ctrl+Shift+F5               # Multiple modifiers
```

### Debugging VHS Scripts

#### 1. Add Debug Screenshots

```bash
# Take screenshots at each step to debug
Type "./app"
Enter
Sleep 2s
Screenshot demo/debug-1-startup.txt

Down Down
Screenshot demo/debug-2-navigation.txt

Space
Screenshot demo/debug-3-selection.txt
```

#### 2. Use Verbose Output

```bash
# Run VHS with debug output
vhs --verbose demo/debug-script.tape
```

#### 3. Test Components Separately

```bash
# Break complex scripts into smaller parts
# test-part-1.tape - Just startup and basic navigation
# test-part-2.tape - Just the problematic feature
# test-part-3.tape - Just cleanup and exit
```

---

## Integration with Development Workflow

### 1. Automated Testing Pipeline

```bash
#!/bin/bash
# ci-test-ui.sh

echo "Building application..."
go build -o filepicker .

echo "Setting up test environment..."
./demo/setup-test-env.sh

echo "Running UI validation tests..."
for test in demo/testing/*.tape; do
    echo "Running $test..."
    vhs "$test"
    
    # Check if test passed (customize validation logic)
    if [ $? -eq 0 ]; then
        echo "✅ $test passed"
    else
        echo "❌ $test failed"
        exit 1
    fi
done

echo "All UI tests passed!"
```

### 2. Documentation Generation

```bash
#!/bin/bash
# generate-docs.sh

echo "Generating documentation GIFs..."

# Build latest version
go build -o filepicker .

# Generate demo GIFs
vhs demo/documentation/overview.tape
vhs demo/documentation/features.tape
vhs demo/documentation/advanced-usage.tape

echo "Documentation GIFs generated:"
ls -la demo/documentation/*.gif
```

### 3. Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: ui-tests
        name: UI Tests
        entry: ./scripts/run-ui-tests.sh
        language: script
        pass_filenames: false
```

### 4. CI/CD Integration

```yaml
# .github/workflows/ui-tests.yml
name: UI Tests
on: [push, pull_request]

jobs:
  ui-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install VHS
        run: |
          curl -s https://api.github.com/repos/charmbracelet/vhs/releases/latest \
          | grep "browser_download_url.*linux_amd64" \
          | cut -d : -f 2,3 \
          | tr -d \" \
          | wget -qi -
          
      - name: Build application
        run: go build -o filepicker .
        
      - name: Run UI tests
        run: ./scripts/run-ui-tests.sh
        
      - name: Upload test artifacts
        uses: actions/upload-artifact@v2
        with:
          name: ui-test-results
          path: demo/testing/*.txt
```

---

## Advanced Patterns

### 1. Parameterized Testing

```bash
#!/bin/bash
# run-responsive-tests.sh

widths=(600 800 1000 1200 1600)

for width in "${widths[@]}"; do
    # Create temporary VHS script
    cat > temp-responsive-test.tape << EOF
Output demo/responsive-${width}.txt
Set Width $width
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s
Screenshot demo/layout-${width}.txt
Type "q"
EOF
    
    vhs temp-responsive-test.tape
    rm temp-responsive-test.tape
done
```

### 2. Data-Driven Testing

```bash
#!/bin/bash
# test-file-types.sh

file_types=("txt" "md" "json" "csv" "jpg" "png")

for type in "${file_types[@]}"; do
    # Create test file
    touch "test-files/sample.$type"
    
    # Test file type detection
    cat > temp-filetype-test.tape << EOF
Output demo/filetype-${type}.txt
Set Width 1200
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to the test file
# (Add navigation logic here)

Screenshot demo/filetype-${type}-display.txt
Type "q"
EOF
    
    vhs temp-filetype-test.tape
    rm temp-filetype-test.tape
    
    # Validate correct icon appears
    if grep -q "sample.$type" "demo/filetype-${type}-display.txt"; then
        echo "✅ File type $type detected correctly"
    else
        echo "❌ File type $type detection failed"
    fi
done
```

### 3. Progressive Feature Testing

```bash
# demo/test-feature-progression.tape
# Tests features in logical progression

Output demo/feature-progression.txt
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Tier 1: Basic navigation
Screenshot demo/tier1-basic.txt

# Tier 2: Enhanced navigation with icons
Down Down
Screenshot demo/tier2-enhanced.txt

# Tier 3: Multi-selection
Space
Down
Space
Screenshot demo/tier3-multi-selection.txt

# Tier 4: Preview panel
Tab
Screenshot demo/tier4-preview.txt

Type "q"
```

---

## Conclusion

VHS is an incredibly powerful tool for TUI development that provides:

- **Reliable Testing**: Automated, reproducible tests for complex UI interactions
- **Beautiful Documentation**: Professional GIFs for README files and documentation
- **Development Confidence**: Catch regressions early with automated validation
- **Visual Debugging**: Screenshot capabilities for understanding UI state

### Key Takeaways

1. **Start Simple**: Begin with basic scripts and gradually add complexity
2. **Test Early**: Integrate VHS testing from the beginning of development
3. **Document Visually**: Use GIFs to show features in action
4. **Validate Continuously**: Screenshot-based validation catches UI regressions
5. **Automate Everything**: Build VHS into your CI/CD pipeline

### Next Steps

1. Set up VHS in your project with the recommended directory structure
2. Create basic validation scripts for your core features
3. Build documentation GIFs for your README
4. Integrate VHS testing into your development workflow
5. Expand to cover edge cases and error conditions

VHS transforms TUI development from a manual, error-prone process into a automated, reliable, and well-documented workflow. It's an essential tool for any serious TUI application development.

---

*This guide was created based on real-world experience developing and testing the Bubbletea File Picker, demonstrating advanced TUI functionality including multi-selection, file operations, responsive layouts, and browser-style navigation history.*
