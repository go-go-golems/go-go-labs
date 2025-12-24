# Changelog

## 2025-12-24

- Initial ticket workspace created with analysis document and task list

## 2025-12-24

Completed full implementation of animal website experiment:
- Created experiment structure under `cmd/experiments/animal-website/`
- Implemented SQLite database with schema migration (`internal/db/`)
- Created animals repository with List, Clear, and InsertMany methods (`internal/animals/`)
- Implemented CSV parsing that extracts names from first column
- Created HTTP handlers with htmx support (`internal/httpui/`)
- Built templ templates for layout, animals list, and upload form (`internal/ui/`)
- Added Bootstrap styling and static file serving
- Application builds successfully and is ready for testing

Key decisions:
- Used `modernc.org/sqlite` (pure Go, no CGO) as specified in analysis
- Default import mode is "replace" (CSV becomes source of truth)
- htmx provides progressive enhancement (works without JS)
- Case-sensitive uniqueness (preserves original case)

Known issues:
- templ generator sometimes adds unused imports (manually removed)
- Upload form uses standard POST (not htmx) for simplicity


