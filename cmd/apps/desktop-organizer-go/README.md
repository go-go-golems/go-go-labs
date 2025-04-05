# desktop-organizer-go

A powerful Go-based tool to analyze directories (like Downloads folders) and generate rich, structured reports about the files within. This tool helps you understand file distribution, find duplicates, identify large files, and get insights that make organizing your directories easier.

## Features

- **Deep file analysis** with intelligent file type detection (using Magika if available)
- **Duplicate file detection** with wasted space calculations
- **Timeline analysis** of when files were created/modified
- **Large file identification** to reclaim disk space
- **Recent files reporting** to see what's been added lately
- **Concurrent processing** for fast analysis even with large directories
- **Multiple output formats** (JSON, with text and markdown support planned)
- **Sampling support** to limit analysis depth for very large directories
- **Path exclusion** via glob patterns

## Installation

### Prerequisites

- Go 1.19 or later
- Optional external tools for enhanced analysis:
  - [Magika](https://github.com/google/magika) - For AI-powered file type detection
  - [ExifTool](https://exiftool.org/) - For detailed metadata extraction from media files
  - [jdupes](https://github.com/jbruchon/jdupes) - For fast duplicate file detection

### Install from source

```bash
git clone https://github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go.git
cd desktop-organizer-go
go build -o desktop-organizer ./cmd/desktop-organizer
```

## Usage

### Basic Scan

Analyze the `~/Downloads` directory and output a report to the console (default format is text):

```bash
./desktop-organizer -d ~/Downloads
```

### Example Scenarios

Here are some common ways to use `desktop-organizer`:

1.  **Quick Scan for Large Files (>500MB):**
    Quickly identify very large files to free up space. Output in JSON format.
    ```bash
    ./desktop-organizer -d /path/to/large-dir --large-file-mb 500 --output-format json -o large_files_report.json
    ```

2.  **Analyze Recent Activity (Last 7 Days) with Debug Logging:**
    See what files have been added recently and get detailed logs.
    ```bash
    ./desktop-organizer -d ~/Downloads --recent-days 7 --log-level debug
    ```

3.  **Detailed Analysis with Sampling, Excluding Caches, Saving Report:**
    Perform a more thorough analysis on a large directory, but sample only 10 files per subdirectory, exclude common cache/temporary folders, and save the report.
    ```bash
    ./desktop-organizer -d /mnt/data \
        -s 10 \
        --exclude-path "*.tmp" \
        --exclude-path ".cache/*" \
        --exclude-path "node_modules/*" \
        -o detailed_report.json \
        --output-format json
    ```

4.  **Specify Custom Tool Path for Magika and Use More Workers:**
    If `magika` isn't in the standard PATH, specify it directly and increase concurrency.
    ```bash
    ./desktop-organizer -d ~/Documents --tool-path magika=/opt/magika/bin/magika --max-workers 12
    ```

5.  **Generate Text Report and Pipe to `less`:**
    Generate the default text report and view it page by page.
    ```bash
    ./desktop-organizer -d /media/external_drive | less
    ```

6.  **Using a Config File but Overriding Output:**
    Use settings from the default config file but force output to stdout.
    ```bash
    # Assuming ~/.desktop-organizer.yaml exists
    ./desktop-organizer -o -
    ```

### All Available Options

The tool offers several flags to customize its behavior. You can combine these flags as needed.

```
Flags:
      --config string            Config file (default is $HOME/.desktop-organizer.yaml or ./config.yaml)
  -d, --downloads-dir string     Directory to analyze (required)
      --debug-log string         Path to write debug logs to a file (JSON format, always debug level)
      --disable-analyzer strings Explicitly disable specific analyzers
      --enable-analyzer strings  Explicitly enable specific analyzers
      --exclude-path strings     Glob patterns for paths to exclude (can specify multiple)
  -h, --help                     Help for desktop-organizer
      --large-file-mb int        Threshold in MB to tag files as 'large' (default: 100)
      --log-level string         Set the logging level (trace, debug, info, warn, error) (default: "info")
      --max-workers int          Number of concurrent workers for file analysis (default: 4)
  -o, --output-file string       Output file path ('-' for stdout) (default: stdout)
      --output-format string     Output format: text, json (default: "text")
      --recent-days int          Threshold in days to tag files as 'recent' (default: 30)
  -s, --sample-per-dir int       Enable sampling: max N files per directory for type analysis (0=disabled)
      --tool-path strings        Override path for external tools (e.g., --tool-path magika=/usr/local/bin/magika)
```

## Configuration

In addition to command-line flags, you can use a YAML configuration file. By default, the tool looks for `.desktop-organizer.yaml` in your home directory or the current directory.

You can specify a custom config path with the `--config` flag:

```bash
./desktop-organizer --config my-config.yaml
```

### Example Configuration File

```yaml
# Target directory to analyze
targetDir: /home/user/Downloads

# Output configuration
outputFile: ~/download-report.json
outputFormat: json

# Analysis settings
samplingPerDir: 20
maxWorkers: 8
largeFileThreshold: 200  # MB
recentFileDays: 14

# Paths to exclude (glob patterns)
excludePaths:
  - "*.tmp"
  - ".git/*"
  - "node_modules/*"

# Override tool paths
toolPaths:
  magika: /usr/local/bin/magika
  exiftool: /usr/bin/exiftool

# Enable/disable specific analyzers
enabledAnalyzers: []  # Empty means use all available
disabledAnalyzers:
  - JdupesDuplicateAnalyzer  # Example if you want to disable a specific analyzer
```

## How It Works

The tool runs a multi-phase analysis pipeline:

1. **Phase 1: File Discovery** - Walks the directory tree, gathers basic info about files.
2. **Phase 2: File Analysis** - Uses specialized analyzers to identify file types, extract metadata, and calculate hashes.
3. **Phase 3: Aggregation** - Finds patterns, groups, duplicates, and generates summary statistics.
4. **Reporting** - Outputs the results in the selected format.

## Example Output

The tool generates structured JSON output that looks similar to:

```json
{
  "root_dir": "/home/user/Downloads",
  "scan_start_time": "2023-07-15T10:30:45Z",
  "scan_end_time": "2023-07-15T10:31:12Z",
  "total_files": 1458,
  "total_dirs": 78,
  "total_size": 4573892544,
  "type_summary": {
    "jpeg": {
      "label": "jpeg",
      "count": 234,
      "size": 567892541
    },
    "pdf": {
      "label": "pdf",
      "count": 123,
      "size": 986543210
    },
    // Other types...
  },
  "duplicate_sets": [
    {
      "id": "3f8d7ae5fdb8c876",
      "file_paths": [
        "photos/vacation1.jpg",
        "backup/vacation1.jpg"
      ],
      "size": 2345678,
      "count": 2,
      "wasted_space": 2345678
    }
    // Other duplicates...
  ],
  "monthly_summary": {
    "2023-06": {
      "year_month": "2023-06",
      "count": 45,
      "size": 128975642
    }
    // Other months...
  }
}
```

## Extending the Tool

The tool is designed to be modular and extensible.

### Adding Custom Analyzers

You can create your own analyzers by implementing the `analysis.Analyzer` interface:

```go
type Analyzer interface {
    Name() string
    Type() AnalyzerType
    DependsOn() []string
    Analyze(ctx context.Context, cfg *config.Config, result *AnalysisResult, entry *FileEntry) error
}
```

### Adding New Output Formats

You can create custom reporters by implementing the `reporting.Reporter` interface:

```go
type Reporter interface {
    FormatName() string
    GenerateReport(ctx context.Context, result *analysis.AnalysisResult, writer io.Writer) error
}
```

## Troubleshooting

### Common Issues

- **"Tool X not available"** - Install the missing external tool or specify a custom path with `--tool-path`
- **Slow Analysis** - Try using sampling (`-s`) for large directories or increase worker count with `--max-workers`
- **High Memory Usage** - Consider using sampling and excluding large subdirectories

### Debug Logs

Enable verbose mode with `-v` and optionally save debug logs to a file:

```bash
./desktop-organizer -d ~/Downloads -v --debug-log debug.log
```

## Contributing

Contributions are welcome! 

1. Fork the repository
2. Create a feature branch
3. Add your changes
4. Submit a pull request

## License

MIT

## Acknowledgments

This project is inspired by and aims to replace the shell script version (`01-inspect-downloads-folder.sh`).
