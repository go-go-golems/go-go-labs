# Quick Start Guide

## Running the Application

### Option 1: Development Mode (Recommended for development)

**Terminal 1 - Start the Go backend:**
```bash
cd /path/to/go-go-labs
go run cmd/apps/pr-history-code-browser/main.go \
  --db /home/manuel/workspaces/2025-10-16/add-gpt5-responses-to-geppetto/geppetto/ttmp/2025-10-23/git-history-and-code-index.db \
  --dev \
  --port 8080
```

**Terminal 2 - Start the Vite dev server:**
```bash
cd cmd/apps/pr-history-code-browser/frontend
npm run dev
```

Then open http://localhost:5173 in your browser.

### Option 2: Production Build

```bash
cd cmd/apps/pr-history-code-browser

# Build frontend
cd frontend
npm install
npm run build
cd ..

# Build Go binary
go build -o pr-history-code-browser main.go

# Run
./pr-history-code-browser \
  --db /home/manuel/workspaces/2025-10-16/add-gpt5-responses-to-geppetto/geppetto/ttmp/2025-10-23/git-history-and-code-index.db \
  --port 8080
```

Then open http://localhost:8080 in your browser.

### Option 3: Using Make

```bash
cd cmd/apps/pr-history-code-browser

# Setup (first time only)
make setup

# Development mode - run these in separate terminals
make dev            # Terminal 1: Backend
make frontend-dev   # Terminal 2: Frontend

# OR build and run production
make build
make run
```

## Features

Once running, you can:

- **Home** - View repository statistics and overview
- **Commits** - Browse all commits with search functionality
  - Click any commit to see details, changed files, and symbols
- **PRs** - View all PR slices with their status
  - Click any PR to see its changelog and analysis notes
- **Files** - Browse tracked files with path prefix filtering
- **Notes** - View analysis notes with tag filtering

## Troubleshooting

### "Database not found"
Make sure the `--db` path points to your `git-history-and-code-index.db` file.

### Frontend shows "Build Frontend First"
Run `cd frontend && npm run build` before starting the backend.

### Port already in use
Change the port: `--port 8081`

### TypeScript/npm errors
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

## Next Steps

See [README.md](README.md) for complete documentation including:
- API endpoints
- Database schema
- Development workflow
- Build scripts

