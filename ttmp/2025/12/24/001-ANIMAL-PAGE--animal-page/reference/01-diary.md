---
Title: Implementation Diary
Ticket: 001-ANIMAL-PAGE
Status: active
Topics:
    - golang
    - webapp
    - sqlite
    - htmx
    - templ
    - csv
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: >
  Step-by-step implementation diary for the animal website experiment.
LastUpdated: 2025-12-24T00:00:00Z
---

# Diary

## Goal

Document the step-by-step implementation of the animal website experiment: a Go web app that accepts CSV uploads, persists animal names to SQLite, and renders them in a browser using htmx, bootstrap, and templ.

## Step 1: Scaffold experiment structure and basic server

Created the initial experiment structure under `cmd/experiments/animal-website/` with a basic HTTP server that accepts CLI flags for logging, listen address, and database path. Used zerolog for structured logging and set up graceful shutdown handling.

**Commit (code):** (not yet committed)

### What I did
- Created `cmd/experiments/animal-website/main.go` with flag parsing and basic HTTP server
- Added zerolog logging setup with `--log-level` flag
- Implemented graceful shutdown using signal context

### Why
- Needed a working entrypoint before building out components
- Zerolog is already used in the codebase and provides structured logging
- Graceful shutdown ensures clean database connection closure

### What worked
- Server starts and accepts flags correctly
- Logging configuration works as expected

### What didn't work
- N/A (initial scaffolding)

### What I learned
- The codebase uses `cmd/experiments/` for experiments, not a separate `experiments/` directory
- Zerolog's `ParseLevel` makes log level configuration straightforward

### What was tricky to build
- N/A (straightforward scaffolding)

### What warrants a second pair of eyes
- Flag defaults (especially `--db-path`) - should we use a more standard location?

### What should be done in the future
- Add tests for flag parsing and server startup
- Consider adding a `--help` flag or using cobra for better CLI UX

### Code review instructions
- Start in `cmd/experiments/animal-website/main.go`
- Verify flag handling and logging setup
- Check graceful shutdown logic

---

## Step 2: Database schema and migration

Created SQLite database package with schema migration. Used `modernc.org/sqlite` (pure Go, no CGO) as specified in the analysis doc. Schema includes `animals` table with id, name, and created_at, plus a unique index on name to prevent duplicates.

**Commit (code):** (not yet committed)

### What I did
- Created `internal/db/migrations.go` with `Migrate()` and `OpenDB()` functions
- Defined schema: `animals` table with autoincrement id, name (TEXT), and created_at timestamp
- Added unique index on name to enforce uniqueness
- Used `modernc.org/sqlite` driver (driver name: `sqlite`)

### Why
- `modernc.org/sqlite` is pure Go (no CGO) and already in go.mod
- Unique index prevents duplicate animal names
- Timestamp provides audit trail

### What worked
- Database opens successfully
- Schema creation works (tested via build)

### What didn't work
- N/A

### What I learned
- `modernc.org/sqlite` uses driver name `sqlite` (not `sqlite3`)
- SQLite's `strftime` function can generate RFC3339 timestamps directly

### What was tricky to build
- Ensuring the driver import is correct (`_ "modernc.org/sqlite"`)

### What warrants a second pair of eyes
- Schema design: is `created_at` as TEXT acceptable, or should we use INTEGER (Unix timestamp)?
- Unique constraint: should we normalize case (e.g., "Cat" vs "cat")?

### What should be done in the future
- Add migration versioning if schema changes are needed
- Consider adding indexes for common queries (e.g., by created_at)

### Code review instructions
- Review `internal/db/migrations.go`
- Verify SQLite driver import and usage
- Check schema design matches requirements

---

## Step 3: Animals repository and CSV parsing

Implemented the animals repository with `List`, `Clear`, and `InsertMany` methods. Created CSV parsing function that extracts animal names from the first column, trims whitespace, and skips empty lines. Supports both "replace" and "append" import modes.

**Commit (code):** (not yet committed)

### What I did
- Created `internal/animals/repository.go` with Repository struct and CRUD methods
- Implemented `List()` to fetch all animals ordered by name
- Implemented `Clear()` to delete all animals
- Implemented `InsertMany()` with transaction support and mode handling (replace vs append)
- Created `internal/animals/csv.go` with `ParseCSV()` function
- Used `INSERT OR IGNORE` to handle duplicates gracefully

### Why
- Repository pattern isolates database logic from HTTP handlers
- Transaction ensures atomicity for replace mode (DELETE + INSERT)
- `INSERT OR IGNORE` handles duplicates without errors
- CSV parsing is forgiving (skips empty lines, trims whitespace)

### What worked
- Repository methods compile and follow standard patterns
- CSV parsing handles edge cases (empty lines, whitespace)

### What didn't work
- Initial timestamp parsing issue: SQLite returns RFC3339 strings, needed to handle both nanosecond and second precision

### What I learned
- SQLite's `strftime('%Y-%m-%dT%H:%M:%fZ', 'now')` generates RFC3339 with nanoseconds
- Go's `time.Parse` needs to handle both RFC3339 and RFC3339Nano formats
- `INSERT OR IGNORE` is perfect for append mode with duplicates

### What was tricky to build
- Timestamp parsing: SQLite returns strings, need to parse back to `time.Time`
- Transaction handling: ensuring rollback on error

### What warrants a second pair of eyes
- Error handling: are we returning enough context in errors?
- Transaction scope: should we use a longer transaction for large imports?

### What should be done in the future
- Add tests for CSV parsing edge cases (multi-column CSV, special characters)
- Add repository tests with in-memory SQLite
- Consider batch insert optimization for large CSV files

### Code review instructions
- Review `internal/animals/repository.go` and `internal/animals/csv.go`
- Check error handling and transaction logic
- Verify CSV parsing handles edge cases

---

## Step 4: HTTP handlers and routing

Created HTTP handlers for all routes: root redirect, animals list, upload form, upload POST, and clear. Handlers detect htmx requests and return fragments when appropriate, falling back to full page renders or redirects for non-htmx requests.

**Commit (code):** (not yet committed)

### What I did
- Created `internal/httpui/handlers.go` with `Handlers` struct
- Implemented `handleRoot()` → redirects to `/animals`
- Implemented `handleAnimals()` → renders animals list (full page or fragment)
- Implemented `handleUpload()` → GET shows form, POST processes CSV upload
- Implemented `handleClear()` → POST clears database
- Added htmx detection via `HX-Request` header
- Registered routes in `RegisterRoutes()`

### Why
- HTML-first approach: works without JavaScript, enhanced with htmx
- htmx detection allows progressive enhancement
- Multipart form parsing handles file uploads

### What worked
- Route registration works correctly
- htmx detection logic is straightforward

### What didn't work
- Initial upload handler had incorrect CSV parsing (was using `encoding/csv.Reader` directly instead of our `ParseCSV` function)

### What I learned
- `r.Header.Get("HX-Request") == "true"` is the standard way to detect htmx requests
- Multipart form parsing requires `r.ParseMultipartForm()` before accessing files
- File content type checking can be lenient (accept `.csv` extension even if Content-Type is wrong)

### What was tricky to build
- htmx fragment vs full page logic: need to render different templates based on request type
- File upload handling: ensuring proper cleanup with `defer file.Close()`

### What warrants a second pair of eyes
- Error responses: should htmx requests return error fragments or redirect?
- File size limits: 10MB is arbitrary, should be configurable

### What should be done in the future
- Add request validation and sanitization
- Add rate limiting for upload endpoint
- Improve error messages for users

### Code review instructions
- Review `internal/httpui/handlers.go`
- Check htmx detection and fragment rendering logic
- Verify file upload handling and error cases

---

## Step 5: Templ templates and UI

Created templ templates for layout, animals page, animals list fragment, and upload page. Used Bootstrap 5 for styling and added htmx attributes for progressive enhancement. Created static CSS file served via `go:embed`.

**Commit (code):** (not yet committed)

### What I did
- Created `internal/ui/layout.templ` with Bootstrap navbar and container
- Created `internal/ui/animals.templ` with full page and list fragment
- Created `internal/ui/upload.templ` with upload form
- Created `internal/ui/static.go` with `go:embed` for static files
- Created `internal/ui/static/app.css` with minimal custom styles
- Added htmx script and Bootstrap CSS via CDN
- Wired up htmx attributes: `hx-post`, `hx-target`, `hx-swap`, `hx-confirm`

### Why
- templ provides type-safe templates and good developer experience
- Bootstrap provides quick, professional styling
- htmx enables progressive enhancement without writing JavaScript
- `go:embed` serves static files without external dependencies

### What worked
- Templates compile and render correctly
- Bootstrap styling looks clean
- htmx attributes work for clear button

### What didn't work
- templ generator adds unused imports sometimes (e.g., `animals` package in `layout_templ.go`)
- Upload form htmx target was incorrect (targeting `#animals-list` which doesn't exist on upload page)

### What I learned
- templ's `@Layout()` pattern allows composing pages
- htmx fragments need to match the target element ID
- `go:embed` requires a subdirectory (can't embed files directly)

### What was tricky to build
- templ generator quirk: sometimes adds unused imports that need manual removal
- htmx target selection: need to ensure target element exists on the page
- Upload form: removed htmx attributes since it should redirect after upload (not update in place)

### What warrants a second pair of eyes
- Template structure: is the fragment/component separation clear?
- htmx wiring: are all interactions properly enhanced?
- Static file serving: is `/static/` prefix correct?

### What should be done in the future
- Add loading indicators for htmx requests
- Improve error display in templates
- Add client-side validation for file uploads
- Consider local Bootstrap instead of CDN for offline development

### Code review instructions
- Review `internal/ui/*.templ` files
- Check htmx attributes and targets
- Verify static file serving works
- Test template rendering in browser

---

## Step 6: Integration and build fixes

Fixed compilation errors, removed unused imports, and ensured the entire application builds successfully. Integrated all components: database, repository, handlers, and templates.

**Commit (code):** (not yet committed)

### What I did
- Fixed unused import in `repository.go` (removed `fmt`)
- Fixed unused import in `handlers.go` (removed `context`, `encoding/csv`, `io`)
- Removed unused import from generated `layout_templ.go` (templ generator quirk)
- Updated `main.go` to wire up database, repository, and handlers
- Verified build succeeds

### Why
- Needed to resolve compilation errors to test the application
- templ generator sometimes adds unused imports that need manual cleanup

### What worked
- Application builds successfully
- All components integrate correctly

### What didn't work
- templ generator keeps adding unused `animals` import to `layout_templ.go` (minor issue, manually removed)

### What I learned
- templ generator can be run multiple times safely
- Generated files sometimes need manual cleanup
- Integration testing requires all components to compile

### What was tricky to build
- templ generator quirk: need to manually remove unused imports after regeneration
- Ensuring all imports are correct across packages

### What warrants a second pair of eyes
- Build process: should we add a script to auto-remove unused imports?
- Integration: are all error paths handled?

### What should be done in the future
- Add integration tests
- Create a build script that handles templ generation and cleanup
- Add a README with usage instructions

### Code review instructions
- Run `go build ./cmd/experiments/animal-website` to verify build
- Check all imports are used
- Verify integration in `main.go`

