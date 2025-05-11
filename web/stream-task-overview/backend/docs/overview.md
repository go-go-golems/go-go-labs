# Stream Task Overview Backend

## Architecture Overview

The Stream Task Overview backend is a Go application built with Echo web framework and SQLite for persistence. It provides a REST API for tracking and managing live stream tasks and information.

## Components

### Main

The entry point of the application which configures and starts the HTTP server. It sets up logging, middleware, routes, and graceful shutdown.

**Key Features:**
- Zerolog structured logging with console output formatting
- Request ID middleware for request tracking
- CORS support for cross-origin requests
- Custom logging middleware for HTTP requests/responses
- Graceful shutdown handling

### Models

Defines the data structures used throughout the application:

- `StreamInfo`: Contains metadata about the stream (title, description, language, etc.)
- `StepInfo`: Contains task steps (completed, active, upcoming)
- `Stream`: Combines StreamInfo and StepInfo into a complete stream state

### Store

The data access layer responsible for persisting and retrieving stream data from SQLite.

**Key Features:**
- Connection management with SQLx
- SQL query building with Squirrel
- Schema initialization
- Default data creation
- Thread-safe data operations with mutex locks
- JSON serialization for array data

### Handlers

Implements HTTP request handlers for the REST API endpoints.

**Key Features:**
- Stream info management (get/update)
- Task steps management (get/set active/complete/add upcoming/reactivate)
- Request validation
- Response formatting
- Error handling

## Database Schema

The application uses SQLite with two main tables:

1. **stream_info**: Stores a single record with stream metadata
2. **steps**: Stores a single record with steps data (arrays are stored as JSON strings)

## Configuration

The application has minimal configuration with hardcoded values:

- Server port: 8080
- Database path: ./stream.db
- Log level: Info (Debug in development mode)

## Development

### Building

```
go build
```

### Running

```
./stream-task-overview
```

Set the ENV variable to "development" for debug logs:

```
ENV=development ./stream-task-overview
```

## API Documentation

See [API Documentation](api.md) for details on the REST endpoints.