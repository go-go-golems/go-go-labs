# LumonStream

A Severance-inspired web application for live coding streamers to display information about their stream, track progress, and manage tasks. Built with Go, React, RTK-Query, SQLite, and Bun.

![LumonStream Screenshot](https://i.imgur.com/example.png)

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Backend Setup](#backend-setup)
  - [Frontend Setup](#frontend-setup)
  - [CLI Setup](#cli-setup)
- [Running the Application](#running-the-application)
  - [Development Mode](#development-mode)
  - [Production Mode](#production-mode)
- [API Documentation](#api-documentation)
- [CLI Usage](#cli-usage)
- [Project Structure](#project-structure)
- [Build System](#build-system)
- [Embedding Frontend in Go Binary](#embedding-frontend-in-go-binary)
- [Troubleshooting](#troubleshooting)
- [Future Enhancements](#future-enhancements)
- [License](#license)

## Features

- **Stream Information Display**: Show title, description, language, GitHub repo, and viewer count
- **Task Management**: Track completed, active, and upcoming tasks
- **Real-time Updates**: Auto-refresh functionality to keep data current
- **Severance-inspired UI**: Clean, corporate aesthetic based on the TV show
- **CLI Tool**: Command-line interface for interacting with the server
- **Single Binary Deployment**: Frontend embedded in Go binary for easy distribution

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go** (version 1.18 or later)
  ```bash
  # Check Go version
  go version
  
  # Install Go (Ubuntu)
  sudo apt-get update
  sudo apt-get install golang
  
  # Install Go (macOS with Homebrew)
  brew install go
  
  # Install Go (Windows)
  # Download from https://golang.org/dl/
  ```

- **Node.js** (version 16 or later)
  ```bash
  # Check Node.js version
  node -v
  
  # Install Node.js (Ubuntu)
  curl -fsSL https://deb.nodesource.com/setup_16.x | sudo -E bash -
  sudo apt-get install -y nodejs
  
  # Install Node.js (macOS with Homebrew)
  brew install node
  
  # Install Node.js (Windows)
  # Download from https://nodejs.org/
  ```

- **Bun** (latest version)
  ```bash
  # Install Bun
  curl -fsSL https://bun.sh/install | bash
  
  # Verify installation
  bun --version
  ```

- **SQLite** (version 3 or later)
  ```bash
  # Install SQLite (Ubuntu)
  sudo apt-get install sqlite3
  
  # Install SQLite (macOS with Homebrew)
  brew install sqlite
  
  # Install SQLite (Windows)
  # Download from https://www.sqlite.org/download.html
  ```

## Installation

### Clone the Repository

```bash
git clone https://github.com/yourusername/LumonStream.git
cd LumonStream
```

### Backend Setup

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Initialize Go modules (if not already done):
   ```bash
   go mod init github.com/lumonstream/backend
   ```

3. Install Go dependencies:
   ```bash
   go get github.com/mattn/go-sqlite3 github.com/gorilla/mux github.com/rs/cors
   ```

4. Build the backend:
   ```bash
   go build -o lumonstream
   ```

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd ../frontend
   ```

2. Install Node.js dependencies:
   ```bash
   npm install
   ```

3. Configure Tailwind CSS (if not already done):
   ```bash
   npx tailwindcss init -p
   ```

### CLI Setup

1. Navigate to the CLI directory:
   ```bash
   cd ../cli
   ```

2. Initialize Go modules (if not already done):
   ```bash
   go mod init github.com/lumonstream/cli
   ```

3. Install Go dependencies:
   ```bash
   go get github.com/spf13/cobra
   ```

4. Build the CLI:
   ```bash
   go build -o lumonstream-cli
   ```

## Running the Application

### Development Mode

In development mode, the backend and frontend run separately, with the frontend development server proxying API requests to the backend.

1. Start the backend server:
   ```bash
   cd backend
   ./lumonstream --debug
   # Or run from source:
   go run main.go --debug
   ```

2. In a separate terminal, start the frontend development server:
   ```bash
   cd frontend
   npm start
   ```

3. Access the application at http://localhost:3000

### Production Mode

In production mode, the frontend is built and embedded in the Go binary, which serves both the API and the static files.

1. Build the application:
   ```bash
   # From the project root
   cd backend
   chmod +x build.sh
   ./build.sh
   ```

2. Run the binary:
   ```bash
   ./lumonstream
   ```

3. Access the application at http://localhost:8080

### Using the CLI

The CLI tool allows you to interact with the server from the command line.

1. Build the CLI (if not already done):
   ```bash
   cd cli
   go build -o lumonstream-cli
   ```

2. Use the CLI to interact with the server:
   ```bash
   # Get stream information
   ./lumonstream-cli get
   
   # Update stream title
   ./lumonstream-cli update --title "New Stream Title"
   
   # Add a new task
   ./lumonstream-cli task add --content "Implement new feature"
   
   # Check server status
   ./lumonstream-cli server status
   ```

## API Documentation

The API documentation is available in the `api-docs.md` file, which details all available endpoints, request/response formats, and data models.

Key endpoints:

- `GET /api/stream-info`: Get stream information and tasks
- `POST /api/stream-info`: Update stream information
- `POST /api/steps`: Add a new task
- `POST /api/steps/status`: Update a task's status

## CLI Usage

The CLI usage instructions are available in the `cli-docs.md` file, which details all available commands, flags, and usage examples.

Key commands:

- `lumonstream-cli get`: Get stream information
- `lumonstream-cli update`: Update stream information
- `lumonstream-cli task add`: Add a new task
- `lumonstream-cli task update`: Update a task's status
- `lumonstream-cli server status`: Check server status

## Project Structure

```
LumonStream/
├── backend/
│   ├── database/
│   │   └── database.go
│   ├── handlers/
│   │   └── handlers.go
│   ├── models/
│   │   └── stream_info.go
│   ├── embed/
│   │   └── (React build files)
│   ├── embed.go
│   ├── embed_debug.go
│   ├── main.go
│   └── build.sh
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   └── store.js
│   │   ├── components/
│   │   │   └── StreamInfoDisplay.jsx
│   │   ├── features/
│   │   │   └── api/
│   │   │       └── apiSlice.js
│   │   ├── App.js
│   │   └── index.js
│   ├── bunbuild.js
│   ├── devserver.js
│   └── package.json
├── cli/
│   ├── cmd/
│   │   ├── root.go
│   │   ├── get.go
│   │   ├── update.go
│   │   ├── task.go
│   │   └── server.go
│   └── main.go
├── package.json
├── tutorial.md
├── api-docs.md
├── cli-docs.md
└── future-enhancements.md
```

## Build System

The build system uses Bun for React compilation and Go's build system for the backend and CLI.

### Root Package.json

The root `package.json` file contains scripts for development, building, and embedding:

```json
{
  "name": "lumonstream-build",
  "scripts": {
    "dev": "cd frontend && bun run start",
    "build": "cd frontend && bun run build",
    "build:embed": "cd frontend && bun run build && cd ../backend && go build -o lumonstream -tags embed"
  }
}
```

### Frontend Build

The frontend build is configured in `frontend/bunbuild.js`:

```javascript
const { build } = require("bun");

async function buildReact() {
  await build({
    entrypoints: ["./src/index.js"],
    outdir: "./build",
    minify: true,
    target: "browser",
    sourcemap: "external",
  });
  
  console.log("React build completed successfully!");
}

module.exports = { buildReact };
```

### Backend Build

The backend build is configured in `backend/build.sh`:

```bash
#!/bin/bash

# Build the React frontend
cd ../frontend
npm run build

# Copy the build files to the backend embed directory
cp -r build/* ../backend/embed/

# Build the Golang binary with embedded files
cd ../backend
go build -tags embed -o lumonstream

echo "Build completed successfully!"
echo "The binary is located at: $(pwd)/lumonstream"
```

## Embedding Frontend in Go Binary

The application uses Go's embed directive to include the frontend build files in the Go binary. This is implemented with build tags to conditionally include the embedding functionality.

### Production Mode (embed.go)

```go
//go:build embed
// +build embed

package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed embed
var embeddedFiles embed.FS

// SetupStaticFiles configures the router to serve embedded static files
func SetupStaticFiles(r *mux.Router) {
	// Create a filesystem with just the embedded files
	fsys, err := fs.Sub(embeddedFiles, "embed")
	if err != nil {
		panic(err)
	}

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.FS(fsys)))
}
```

### Debug Mode (embed_debug.go)

```go
//go:build !embed
// +build !embed

package main

import (
	"github.com/gorilla/mux"
)

// SetupStaticFiles is a no-op in debug mode
func SetupStaticFiles(r *mux.Router) {
	// In debug mode, we don't serve static files from the backend
	// The frontend development server will handle this
}
```

## Troubleshooting

### Backend Issues

- **Database Errors**: Ensure SQLite is installed and the database file is writable.
  ```bash
  # Check SQLite installation
  sqlite3 --version
  
  # Ensure database directory is writable
  chmod -R 755 /path/to/database/directory
  ```

- **Port Already in Use**: Change the port using the `--port` flag.
  ```bash
  ./lumonstream --port 9000
  ```

### Frontend Issues

- **Node Modules Errors**: Try reinstalling node modules.
  ```bash
  cd frontend
  rm -rf node_modules
  npm install
  ```

- **Build Errors**: Ensure all dependencies are installed.
  ```bash
  cd frontend
  npm install
  ```

### CLI Issues

- **Connection Errors**: Ensure the server is running and the URL is correct.
  ```bash
  # Check server status
  ./lumonstream-cli --server http://localhost:8080 server status
  ```

## Future Enhancements

See the `future-enhancements.md` file for a comprehensive list of potential future enhancements, including:

- Authentication system
- WebSocket support for real-time updates
- Database improvements
- API extensions
- UI improvements
- Additional components
- CLI enhancements
- Build system improvements
- Plugin system
- Integration options
- Analytics and monitoring
- Internationalization
- Content management

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [Go](https://golang.org/)
- [React](https://reactjs.org/)
- [Redux Toolkit](https://redux-toolkit.js.org/)
- [RTK Query](https://redux-toolkit.js.org/rtk-query/overview)
- [Cobra](https://github.com/spf13/cobra)
- [Bun](https://bun.sh/)
- [SQLite](https://www.sqlite.org/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Lucide Icons](https://lucide.dev/)
