# Animal Website

A simple Go web application that accepts CSV uploads of animal names, persists them to SQLite, and displays them in a browser.

## Features

- Upload CSV files with animal names
- View list of animals in a table
- Clear all animals
- Replace or append import modes
- Progressive enhancement with htmx
- Bootstrap styling

## Usage

### Run the server

```bash
go run ./cmd/experiments/animal-website
```

Options:
- `--listen-addr`: Address to listen on (default: `:8080`)
- `--db-path`: Path to SQLite database file (default: `./animals.db`)
- `--log-level`: Log level: debug, info, warn, error (default: `info`)

### CSV Format

The CSV file should contain animal names, one per line. The parser:
- Reads the first column of each row
- Trims whitespace
- Skips empty lines
- Preserves case (case-sensitive uniqueness)

Example CSV:
```csv
cat
dog
capybara
elephant
```

### Routes

- `GET /` → Redirects to `/animals`
- `GET /animals` → Animals list page
- `GET /upload` → Upload form page
- `POST /upload` → Process CSV upload
- `POST /animals/clear` → Clear all animals

## Development

### Generate templ files

```bash
templ generate ./cmd/experiments/animal-website/internal/ui
```

Note: The templ generator sometimes adds unused imports. You may need to manually remove them from generated files.

### Build

```bash
go build ./cmd/experiments/animal-website
```

## Technology Stack

- **Go**: HTTP server and business logic
- **SQLite**: Database (via `modernc.org/sqlite`)
- **templ**: Type-safe HTML templates
- **htmx**: Progressive enhancement
- **Bootstrap 5**: CSS framework

