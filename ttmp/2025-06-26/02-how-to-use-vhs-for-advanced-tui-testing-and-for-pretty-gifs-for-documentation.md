# The Complete Guide to VHS for Advanced TUI Testing and Documentation

**Author**: AI Assistant  
**Date**: Updated September 15, 2025  
**Project**: Advanced TUI Development with Bubbletea File Picker  
**VHS Version**: v0.8.0+

## Table of Contents

1. [Introduction to VHS](#introduction-to-vhs)
2. [Installation and Setup](#installation-and-setup)
3. [Basic VHS Syntax](#basic-vhs-syntax)
4. [Advanced Testing Techniques](#advanced-testing-techniques)
5. [Validation Workflows](#validation-workflows)
6. [Creating Documentation GIFs](#creating-documentation-gifs)
7. [Real-World Examples](#real-world-examples)
8. [Modern Best Practices (2025)](#modern-best-practices-2025)
9. [Container and CI/CD Integration](#container-and-cicd-integration)
10. [Performance Optimization](#performance-optimization)
11. [Troubleshooting](#troubleshooting)
12. [Integration with Development Workflow](#integration-with-development-workflow)

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

**Latest Installation Methods (September 2025):**

```bash
# macOS with Homebrew (recommended)
brew install vhs

# Linux with package managers
# Ubuntu/Debian
curl -fsSL https://charm.sh/vhs/install.sh | bash

# Arch Linux
yay -S vhs

# Or with Go (latest development version)
go install github.com/charmbracelet/vhs@latest

# Verify installation and check features
vhs --version
vhs --help
```

### Install Dependencies

VHS requires additional tools for full functionality:

```bash
# macOS
brew install ttyd imagemagick

# Ubuntu/Debian  
sudo apt update
sudo apt install ttyd imagemagick

# Arch Linux
sudo pacman -S ttyd imagemagick

# Install modern fonts with better Unicode support
# macOS
brew install font-jetbrains-mono-nerd-font font-fira-code-nerd-font

# Linux (manual installation)
wget -P ~/.local/share/fonts \
  'https://github.com/ryanoasis/nerd-fonts/releases/download/v3.0.2/JetBrainsMono.zip' && \
  unzip ~/.local/share/fonts/JetBrainsMono.zip -d ~/.local/share/fonts/ && \
  fc-cache -fv
```

### Modern Container-Based Setup

For CI/CD and reproducible environments:

```dockerfile
# Dockerfile.vhs
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl \
    ttyd \
    imagemagick \
    fonts-jetbrains-mono \
    && rm -rf /var/lib/apt/lists/*

# Install VHS
RUN curl -fsSL https://charm.sh/vhs/install.sh | bash

# Copy your app and VHS scripts
COPY ./your-app /app/your-app
COPY ./demo /app/demo
WORKDIR /app

# Run VHS scripts
CMD ["sh", "-c", "find demo -name '*.tape' -exec vhs {} \\;"]
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

## Modern Best Practices (2025)

### 1. Environment Consistency

**Use Container-Based Workflows:**

```yaml
# .github/workflows/vhs-tests.yml
name: VHS TUI Tests

on: [push, pull_request]

jobs:
  ui-tests:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Build TUI application
        run: |
          go build -o myapp .
          
      - name: Run VHS tests in container
        run: |
          docker build -f Dockerfile.vhs -t vhs-test .
          docker run --rm -v $(pwd)/demo:/app/demo vhs-test
          
      - name: Upload test artifacts
        uses: actions/upload-artifact@v3
        with:
          name: vhs-outputs
          path: |
            demo/**/*.txt
            demo/**/*.gif
            demo/**/*.mp4
```

**VHS Configuration Standards:**

```bash
# vhs.config.tape - Global settings file
Set Shell "zsh"              # Use consistent shell
Set Theme "GitHub Dark"      # Professional theme for docs
Set FontFamily "JetBrains Mono NL"
Set FontSize 13
Set Framerate 60             # Smooth animations
Set PlaybackSpeed 1.0        # Real-time playback
Set Margin 20               # Better framing
Set MarginFill "#1e1e2e"    # Professional background
```

### 2. Advanced Script Organization

**Modular Script Architecture:**

```
demo/
├── config/
│   ├── common.tape          # Shared settings
│   └── themes.tape          # Theme definitions  
├── components/              # Reusable script components
│   ├── startup.tape         # App initialization
│   ├── navigation.tape      # Common navigation patterns
│   └── cleanup.tape         # Consistent exit sequences
├── tests/
│   ├── smoke/              # Quick validation tests
│   ├── integration/        # Full workflow tests
│   └── regression/         # Bug reproduction tests
└── docs/
    ├── features/           # Feature demonstration GIFs
    ├── tutorials/          # Step-by-step guides
    └── troubleshooting/    # Error scenario docs
```

**Script Composition Pattern:**

```bash
# demo/tests/integration/full-workflow.tape
Source config/common.tape
Source components/startup.tape

# Main test logic
Type "specific command"
Sleep 1s

Source components/cleanup.tape
```

### 3. Performance-Optimized Recording

**Smart Timing Strategy:**

```bash
# Use dynamic timing based on operation type
# Fast operations
Type "ls"
Sleep 200ms

# Medium operations (file I/O)  
Type "cat large-file.txt"
Sleep 800ms

# Heavy operations (searching, processing)
Type "/search-term"
Sleep 2s

# UI rendering delays
Type "F3"                    # Toggle view
Sleep 500ms                  # Wait for render
Screenshot demo/view.txt     # Capture result
```

**Optimized Output Settings:**

```bash
# For documentation (balance quality/size)
Output demo.gif
Set Width 1200
Set Height 800
Set Quality 80              # Reduce file size
Set Framerate 30            # Smooth but efficient

# For detailed testing (high fidelity)
Output test.mp4
Set Width 1600
Set Height 1000
Set Quality 100
Set Framerate 60
```

### 4. Advanced Validation Techniques

**Comprehensive Screenshot Validation:**

```bash
#!/bin/bash
# validate-screenshots.sh

# Define expected UI states
declare -A expected_states=(
    ["initial"]="File Browser.*test-files"
    ["selected"]="▶.*file1.txt.*◀"
    ["copied"]="Copied.*2.*files"
    ["error"]="Error:.*Permission denied"
)

# Validate each state
for state in "${!expected_states[@]}"; do
    if grep -qP "${expected_states[$state]}" "demo/${state}.txt"; then
        echo "✅ $state state valid"
    else
        echo "❌ $state state invalid"
        echo "Expected: ${expected_states[$state]}"
        echo "Actual:"
        cat "demo/${state}.txt"
        exit 1
    fi
done
```

**Automated Visual Regression Testing:**

```bash
#!/bin/bash
# visual-regression-check.sh

# Generate current screenshots
vhs demo/baseline.tape

# Compare with reference
for file in demo/screenshots/*.txt; do
    baseline="demo/reference/$(basename "$file")"
    if [ -f "$baseline" ]; then
        if ! diff -q "$file" "$baseline" > /dev/null; then
            echo "❌ Visual regression detected in $(basename "$file")"
            echo "Differences:"
            diff "$baseline" "$file"
            exit 1
        fi
    fi
done

echo "✅ No visual regressions detected"
```

---

## Container and CI/CD Integration

### 1. Docker-Based Testing Pipeline

**Multi-Stage Dockerfile:**

```dockerfile
# Dockerfile.vhs-ci
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp .

FROM ubuntu:22.04 AS vhs-runner
RUN apt-get update && apt-get install -y \
    curl \
    ttyd \
    imagemagick \
    fonts-jetbrains-mono \
    xvfb \
    && rm -rf /var/lib/apt/lists/*

# Install latest VHS
RUN curl -fsSL https://charm.sh/vhs/install.sh | bash

# Create virtual display for headless operation
ENV DISPLAY=:99
RUN mkdir -p /tmp/.X11-unix

COPY --from=builder /app/myapp /app/
COPY demo/ /app/demo/
WORKDIR /app

# Run with virtual display
CMD ["sh", "-c", "Xvfb :99 -screen 0 1920x1080x24 & exec ./run-vhs-tests.sh"]
```

### 2. GitHub Actions Integration

**Comprehensive CI Workflow:**

```yaml
# .github/workflows/vhs-comprehensive.yml
name: Comprehensive VHS Testing

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  vhs-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        test-suite: [smoke, integration, regression]
        
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Build application
        run: go build -o myapp .
        
      - name: Install VHS
        run: |
          curl -fsSL https://charm.sh/vhs/install.sh | bash
          sudo apt-get update
          sudo apt-get install -y ttyd imagemagick fonts-jetbrains-mono
          
      - name: Create test environment
        run: |
          mkdir -p test-data
          ./scripts/setup-test-data.sh
          
      - name: Run VHS test suite
        run: |
          export DISPLAY=:99
          Xvfb :99 -screen 0 1920x1080x24 &
          sleep 2
          ./scripts/run-vhs-suite.sh ${{ matrix.test-suite }}
          
      - name: Validate outputs
        run: ./scripts/validate-vhs-outputs.sh
        
      - name: Upload test results
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: vhs-results-${{ matrix.test-suite }}
          path: |
            demo/**/*.txt
            demo/**/*.gif
            demo/**/*.mp4
            
      - name: Update documentation
        if: github.ref == 'refs/heads/main' && matrix.test-suite == 'integration'
        run: |
          # Copy generated GIFs to docs
          cp demo/docs/**/*.gif docs/assets/
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add docs/assets/
          git commit -m "Update VHS documentation GIFs" || exit 0
          git push
```

### 3. Performance Monitoring

**VHS Performance Profiler:**

```bash
#!/bin/bash
# vhs-performance-profiler.sh

echo "VHS Performance Profile Report"
echo "================================"

for tape in demo/**/*.tape; do
    echo "Profiling: $tape"
    
    start_time=$(date +%s.%N)
    timeout 60s vhs "$tape" || echo "TIMEOUT: $tape"
    end_time=$(date +%s.%N)
    
    duration=$(echo "$end_time - $start_time" | bc)
    echo "Duration: ${duration}s"
    
    # Check output file sizes
    output=$(grep "^Output" "$tape" | head -1 | awk '{print $2}')
    if [ -f "$output" ]; then
        size=$(du -h "$output" | cut -f1)
        echo "Output size: $size"
    fi
    
    echo "---"
done
```

---

## Performance Optimization

### 1. Recording Optimization

**Efficient Recording Settings:**

```bash
# Optimized for CI/CD (faster execution)
Set Framerate 15           # Lower framerate for CI
Set Quality 60             # Balanced quality/speed
Set PlaybackSpeed 1.5      # Faster playback

# Optimized for Documentation (higher quality)
Set Framerate 30
Set Quality 90
Set PlaybackSpeed 1.0
```

**Selective Output Generation:**

```bash
# Conditional output based on environment
{% raw %}
{{- if .CI }}
Output test-results.txt     # Text only in CI
{{- else }}
Output demo.gif            # Full GIF for local dev
Output demo.txt            # Text for validation
{{- end }}
{% endraw %}
```

### 2. Resource Management

**Memory-Efficient Testing:**

```bash
#!/bin/bash
# run-vhs-batched.sh - Process VHS scripts in batches

batch_size=3
scripts=(demo/**/*.tape)

for ((i=0; i<${#scripts[@]}; i+=batch_size)); do
    batch=("${scripts[@]:i:batch_size}")
    
    echo "Processing batch: ${batch[*]}"
    
    # Run batch in parallel with resource limits
    for script in "${batch[@]}"; do
        (
            ulimit -m 512000  # 512MB memory limit
            timeout 60s vhs "$script"
        ) &
    done
    
    wait  # Wait for batch to complete
    
    # Clean up temporary files
    find /tmp -name "vhs-*" -mmin +5 -delete
done
```

### 3. Caching Strategies

**VHS Output Caching:**

```bash
#!/bin/bash
# vhs-with-cache.sh

script="$1"
script_hash=$(shasum -a 256 "$script" | cut -d' ' -f1)
cache_dir="$HOME/.vhs-cache"
cached_output="$cache_dir/$script_hash"

mkdir -p "$cache_dir"

if [ -f "$cached_output" ]; then
    echo "Using cached result for $script"
    cp "$cached_output" "$(grep '^Output' "$script" | awk '{print $2}')"
else
    echo "Running VHS for $script"
    vhs "$script"
    # Cache the result
    output_file=$(grep '^Output' "$script" | awk '{print $2}')
    if [ -f "$output_file" ]; then
        cp "$output_file" "$cached_output"
    fi
fi
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

VHS has evolved into an incredibly powerful and mature tool for TUI development that provides:

- **Reliable Testing**: Automated, reproducible tests for complex UI interactions
- **Beautiful Documentation**: Professional GIFs for README files and documentation  
- **Development Confidence**: Catch regressions early with automated validation
- **Visual Debugging**: Screenshot capabilities for understanding UI state
- **CI/CD Integration**: Seamless integration with modern development workflows
- **Performance Optimization**: Advanced caching and batching for efficient testing

### Modern Development Benefits (2025)

1. **Container-First Approach**: Reproducible testing across all environments
2. **Advanced Validation**: Automated visual regression testing and comprehensive screenshot validation
3. **Performance Monitoring**: Built-in profiling and resource management
4. **Modular Architecture**: Reusable script components and configuration management
5. **Professional Output**: High-quality documentation assets with optimized settings

### Implementation Roadmap

#### Phase 1: Foundation (Week 1)
1. Set up VHS with container-based workflow
2. Create modular script architecture
3. Implement basic validation scripts
4. Configure CI/CD pipeline integration

#### Phase 2: Optimization (Week 2-3)
1. Add performance monitoring and caching
2. Implement advanced validation techniques
3. Create comprehensive test suites (smoke, integration, regression)
4. Optimize output settings for different use cases

#### Phase 3: Advanced Features (Week 4+)
1. Visual regression testing pipeline
2. Automated documentation generation
3. Performance profiling and optimization
4. Custom validation frameworks

### Best Practices Summary

1. **Environment Consistency**: Use containers and standardized configurations
2. **Modular Design**: Break scripts into reusable components
3. **Smart Validation**: Combine screenshot and behavioral testing
4. **Performance First**: Implement caching and resource management
5. **Automate Everything**: Full CI/CD integration with artifact management

### Migration from Legacy VHS Setups

If upgrading from older VHS implementations:

```bash
#!/bin/bash
# migrate-vhs-setup.sh

# 1. Update to latest VHS version
curl -fsSL https://charm.sh/vhs/install.sh | bash

# 2. Reorganize script structure
mkdir -p demo/{config,components,tests/{smoke,integration,regression},docs}

# 3. Convert old scripts to modular format
for script in demo/*.tape; do
    echo "Converting $script to modular format..."
    # Add Source directives, extract common settings
done

# 4. Set up container workflow
cp templates/Dockerfile.vhs-ci ./
cp templates/.github-workflows-vhs.yml .github/workflows/

echo "Migration complete. Review generated files and customize as needed."
```

### Future-Proofing Your VHS Setup

As VHS continues to evolve, ensure your setup remains maintainable:

1. **Version Pinning**: Pin VHS versions in your CI for consistency
2. **Configuration Management**: Use centralized configuration files
3. **Regular Updates**: Schedule quarterly reviews of VHS updates
4. **Community Engagement**: Follow VHS releases and community best practices
5. **Documentation**: Keep internal documentation updated with your specific workflows

VHS has transformed from a simple terminal recording tool into a comprehensive TUI testing and documentation platform. By following the modern practices outlined in this guide, you'll create robust, maintainable, and professional TUI applications with confidence.

---

*This guide reflects real-world experience with VHS across multiple TUI projects as of September 2025, incorporating the latest features, best practices, and integration patterns for modern development workflows.*
