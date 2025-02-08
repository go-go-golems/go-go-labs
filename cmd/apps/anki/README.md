# Anki Viewer

A modern web interface for viewing Anki decks, cards, and note types (models). This application provides a clean, user-friendly way to explore your Anki collection without modifying it.

## Features

- View all Anki decks in a clean list interface
- Explore note types (models) with detailed information:
  - Field definitions and descriptions
  - Card templates (front and back)
  - Styling (CSS)
- View cards within decks with proper HTML rendering
- Debug logging for API communication

## Architecture

The application follows a clean, modular architecture:

```
cmd/apps/anki/
├── main.go           # Entry point and HTTP server
├── services/         # Business logic and Anki communication
│   └── anki.go      # AnkiConnect API client
├── views/           # UI templates
│   ├── layout.templ # Base layout template
│   ├── decks.templ  # Decks list view
│   ├── cards.templ  # Cards view
│   └── models.templ # Models view
└── README.md        # This file
```

### Components

1. **HTTP Server (Echo)**
   - Handles routing and HTTP requests
   - Serves the web interface
   - Manages middleware (logging, recovery)

2. **Anki Service**
   - Communicates with Anki via AnkiConnect
   - Handles API requests and responses
   - Provides type-safe interfaces for Anki operations

3. **View Layer (templ)**
   - Type-safe templates
   - Component-based UI
   - HTML generation with Go integration

## Technologies Used

- **Go** - Backend language and runtime
- **Echo** - HTTP server framework
- **templ** - Type-safe templating engine
- **HTMX** - Dynamic UI updates without JavaScript
- **Bootstrap 5** - Frontend styling and components
- **zerolog** - Structured logging
- **cobra** - CLI command and flags handling

## Communication with Anki

The application communicates with Anki through the [AnkiConnect](https://foosoft.net/projects/anki-connect/) add-on using its HTTP API:

1. **Protocol**
   - HTTP POST requests to `localhost:8765`
   - JSON-encoded requests and responses
   - Version 6 of the AnkiConnect API

2. **Key API Endpoints Used**
   - `deckNames` - Get list of decks
   - `findCards` - Find cards in a deck
   - `cardsInfo` - Get detailed card information
   - `modelNamesAndIds` - Get list of note types
   - `findModelsByName` - Get detailed model information

3. **Example Request/Response**
```json
// Request
{
    "action": "deckNames",
    "version": 6
}

// Response
{
    "result": ["Default", "MyDeck"],
    "error": null
}
```

## Setup and Running

1. **Prerequisites**
   - Go 1.21 or later
   - Anki with AnkiConnect add-on installed
   - Anki running and AnkiConnect active

2. **Installation**
   ```bash
   # Install dependencies
   go mod tidy
   
   # Generate templ templates
   templ generate
   ```

3. **Running**
   ```bash
   # Run with default info logging
   go run main.go

   # Run with debug logging (shows API communication)
   go run main.go --log-level debug
   ```

4. **Access**
   - Open `http://localhost:8080` in your browser
   - Anki must be running with AnkiConnect add-on active

## Development

### Adding New Features

1. **New Anki Operations**
   - Add methods to `AnkiService` in `services/anki.go`
   - Follow existing patterns for error handling and logging

2. **New UI Components**
   - Create new `.templ` files in `views/`
   - Use existing layout and Bootstrap components
   - Add new routes in `main.go`

### Debugging

- Use `--log-level debug` to see full API communication
- All requests and responses are logged in pretty-printed JSON
- HTTP server logs show all incoming requests

## Notes

- The application is read-only and cannot modify your Anki collection
- Anki must be running for the application to work
- Some HTML content from cards may require additional styling
- The application assumes AnkiConnect is available on the default port 