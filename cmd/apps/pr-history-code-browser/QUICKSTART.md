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
  - **NEW:** See which PRs used this commit (cross-references)
  - **NEW:** View analysis notes linked to the commit
- **PRs** - View all PR slices with their status
  - Click any PR to see its changelog and analysis notes
  - **NEW:** Changelog entries show full commit details (clickable hashes)
  - **NEW:** See referenced files for each changelog entry
  - **NEW:** Analysis notes show their related commits and files
- **Files** - Browse tracked files with path prefix filtering
  - **NEW:** Click any file to see its detailed history
  - **NEW:** View files often changed together (co-change analysis)
  - **NEW:** See recent commits and analysis notes for each file
- **Notes** - View analysis notes with tag filtering

### Key Cross-Referencing Features

The app now provides **complete bidirectional cross-linking** between all entities:

1. **Commit ↔ PRs**: See which PRs used each commit (and vice versa)
2. **Commit ↔ Files**: View changed files with clickable links to file details
3. **PR ↔ Files**: See which files were referenced in PR changelog (clickable)
4. **File ↔ PRs**: NEW! See which PRs referenced each file
5. **File ↔ Files**: Co-change analysis shows files often changed together
6. **Notes ↔ Commits/Files**: All note references are clickable
7. **Contextual Navigation**: Click any entity reference to navigate instantly

**Every relationship is clickable!** Navigate seamlessly through your codebase history.

See [CROSS-LINKING-COMPLETE.md](CROSS-LINKING-COMPLETE.md) for complete documentation.

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

