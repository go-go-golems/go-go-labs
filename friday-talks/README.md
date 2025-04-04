# Friday Talks

A web application for scheduling and managing Friday knowledge sharing sessions among friends or colleagues.

## Features

- User authentication system
- Propose, edit, and schedule talks
- Vote on proposed talks to show interest
- Calendar view of scheduled talks
- Resource sharing (slides, videos, code, etc.)
- Track attendance and feedback

## Technology Stack

- **Backend**: Go with Chi router
- **Frontend**: Server-side rendered HTML with [HTMX](https://htmx.org/) and Bootstrap 5
- **Database**: SQLite
- **Authentication**: JWT tokens
- **Templates**: [Templ](https://github.com/a-h/templ) for Go HTML templates

## Installation

### Prerequisites

- Go 1.18 or later
- SQLite

### Building from Source

1. Clone the repository:
```sh
git clone https://github.com/your-username/friday-talks.git
cd friday-talks
```

2. Install dependencies:
```sh
go mod download
```

3. Generate templates:
```sh
go install github.com/a-h/templ/cmd/templ@latest
templ generate
```

4. Build the application:
```sh
go build -o friday-talks ./cmd/server
```

## Usage

1. Run the application:
```sh
./friday-talks
```

2. Open your browser and go to:
```
http://localhost:8080
```

### Command Line Options

- `--port, -p`: Port to listen on (default: 8080)
- `--db, -d`: Path to SQLite database file (default: friday-talks.db)
- `--jwt-secret, -j`: Secret key for JWT tokens (default: your-secret-key)
- `--static, -s`: Path to static files (default: static)

Example:
```sh
./friday-talks --port 9000 --db ./data/talks.db
```

## Development

### Project Structure

```
/friday-talks
  /cmd
    /server        # Main application entry point
  /internal
    /auth          # Authentication logic
    /handlers      # HTTP request handlers
    /models        # Data models and database access
    /templates     # Templ HTML templates
    /services      # Business logic
  /migrations      # Database migrations
  /static          # CSS, JS, images
  /docs            # Documentation
  go.mod           # Go module definition
  README.md        # Project documentation
```

### Database Migrations

The application automatically applies migrations during startup. Migrations are located in the `/migrations` directory.

## Security Considerations

For production use:

1. Change the default JWT secret key
2. Use HTTPS for all connections
3. Store sensitive configuration in environment variables or a secure configuration system
4. Consider using a more robust database like PostgreSQL for larger deployments

## License

This project is licensed under the MIT License - see the LICENSE file for details.