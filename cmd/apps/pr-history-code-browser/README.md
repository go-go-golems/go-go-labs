# PR History & Code Browser

A web application that visualizes the git history database and PR work tracking from the SQLite database created by `build_history_index.py`.

## Overview

This application provides an interactive interface to:
- Browse commit history with search functionality
- View detailed commit information including changed files and symbols
- Explore PR slices with their changelogs and analysis notes
- Search and filter files in the repository
- Read analysis notes with tag filtering

## Architecture

The application consists of two main components:

### Backend (Go)
- **Framework**: Go with `chi` router
- **Database**: SQLite (read-only mode)
- **API**: RESTful JSON API
- **Features**:
  - Statistics endpoint for repository overview
  - Commit listing and detailed views with pagination
  - PR management with changelog and notes
  - File browsing with history tracking
  - Analysis notes with filtering

### Frontend (React + TypeScript)
- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite
- **Routing**: React Router v7
- **Styling**: Custom CSS with modern design
- **Features**:
  - Responsive design with clean UI
  - Search and pagination
  - Detailed views for commits and PRs
  - Real-time filtering

## Prerequisites

- Go 1.24.3 or higher
- Node.js 22 or higher
- SQLite database file from `build_history_index.py`

## Installation

### 1. Install Go dependencies

The application uses the go-go-labs top-level module, so dependencies are already available:

```bash
cd /path/to/go-go-labs
go mod download
```

### 2. Install Frontend dependencies

```bash
cd cmd/apps/pr-history-code-browser/frontend
npm install
```

## Development

### Running in Development Mode

For development, you'll run both the backend and frontend separately:

#### Terminal 1: Start the Go backend

```bash
cd /path/to/go-go-labs
go run cmd/apps/pr-history-code-browser/main.go \
  --db /path/to/git-history-and-code-index.db \
  --dev \
  --port 8080
```

The `--dev` flag enables CORS for the frontend development server.

#### Terminal 2: Start the Vite dev server

```bash
cd cmd/apps/pr-history-code-browser/frontend
npm run dev
```

The frontend will be available at `http://localhost:5173` and will proxy API requests to the Go backend at `http://localhost:8080`.

## Production Build

### 1. Build the Frontend

```bash
cd cmd/apps/pr-history-code-browser/frontend
npm run build
```

This creates optimized static files in `frontend/dist/`.

### 2. Build the Go Binary

```bash
cd /path/to/go-go-labs
go build -o pr-history-code-browser cmd/apps/pr-history-code-browser/main.go
```

### 3. Run the Production Server

```bash
./pr-history-code-browser --db /path/to/git-history-and-code-index.db --port 8080
```

The application will serve both the API and the static frontend at `http://localhost:8080`.

## Usage

### Command Line Options

```bash
pr-history-code-browser [flags]

Flags:
  -d, --db string      Path to SQLite database file (required)
  -p, --port int       Port to listen on (default 8080)
      --dev            Enable development mode (allows CORS)
  -h, --help           Help for pr-history-code-browser
```

### Example with Geppetto Database

```bash
# Development mode
go run cmd/apps/pr-history-code-browser/main.go \
  --db /home/manuel/workspaces/2025-10-16/add-gpt5-responses-to-geppetto/geppetto/ttmp/2025-10-23/git-history-and-code-index.db \
  --dev

# Production mode
./pr-history-code-browser \
  --db /home/manuel/workspaces/2025-10-16/add-gpt5-responses-to-geppetto/geppetto/ttmp/2025-10-23/git-history-and-code-index.db
```

## API Endpoints

### Statistics
- `GET /api/stats` - Get repository statistics

### Commits
- `GET /api/commits?limit=50&offset=0&search=query` - List commits with pagination and search
- `GET /api/commits/:hash` - Get commit details with files and symbols

### Pull Requests
- `GET /api/prs` - List all PRs
- `GET /api/prs/:id` - Get PR details with changelog and notes

### Files
- `GET /api/files?limit=100&offset=0&prefix=path` - List files with filtering
- `GET /api/files/:id/history?limit=50` - Get file history

### Analysis Notes
- `GET /api/notes?limit=50&offset=0&type=manual-review&tags=PR03` - List notes with filtering

## Database Schema

The application reads from these tables:
- `commits` - Git commit metadata
- `files` - Tracked file paths
- `commit_files` - File changes per commit
- `commit_symbols` - Extracted symbols (functions, types, etc.)
- `prs` - PR slice definitions
- `pr_changelog` - PR work tracking
- `analysis_notes` - Manual annotations

See `geppetto/ttmp/2025-10-23/git-history-index-guide.md` for details on database structure and how it's created.

## Features

### Home Page
- Repository statistics overview
- PR status breakdown
- Quick navigation links

### Commits View
- Paginated commit list
- Search by hash, subject, or body
- Click to view detailed commit information

### Commit Detail
- Full commit metadata
- List of changed files with stats
- Symbol extraction results
- Parent commit links

### PRs View
- List of all PR slices
- Status indicators
- Click to view PR details

### PR Detail
- PR description and status
- Complete changelog of actions
- Associated analysis notes
- Tag filtering

### Files View
- Browse all tracked files
- Filter by path prefix
- Pagination support

### Notes View
- Browse analysis notes
- Filter by tags
- Chronological ordering

## Troubleshooting

### Frontend not loading in production
- Ensure you've run `npm run build` in the frontend directory
- Check that `frontend/dist/` exists and contains files
- Run without `--dev` flag

### Database errors
- Verify the database path is correct
- Ensure the database file is readable
- Check that the database was created by `build_history_index.py`

### Port already in use
- Change the port with `--port` flag
- Check for other processes using the port: `lsof -i :8080`

## Related Documentation

- `geppetto/ttmp/2025-10-23/01-how-to-use-the-sqlite-db-to-create-prs.md` - Workflow guide
- `geppetto/ttmp/2025-10-23/git-history-index-guide.md` - Database creation and schema
- `geppetto/ttmp/2025-10-23/feature-history-timeline.md` - Feature timeline narrative
- `geppetto/ttmp/2025-10-23/pr-extraction-guide.md` - PR extraction strategies

## License

Part of the go-go-labs project.

