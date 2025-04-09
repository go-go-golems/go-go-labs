# Messy Log Server Example

This is an intentionally messy implementation of a REST API and web UI for a log server. It's designed to serve as an example of "dirty code" that needs refactoring.

## Features

- REST API for logs, users, metrics, and configuration
- Simple web UI dashboard
- In-memory data storage with global variables
- Hard-coded styles and scripts
- No proper separation of concerns
- Duplicated code
- Inconsistent error handling
- No proper middleware implementation
- No proper routing
- Mixed business logic and presentation

## Running the Server

```bash
cd cmd/experiments/server-logs
go run *.go
```

The server will start on port 8080 by default. You can specify a different port as a command-line argument:

```bash
go run *.go 9000
```

## API Endpoints

### Logs

- `GET /api/logs` - Get all logs
- `GET /api/logs?level=info` - Filter logs by level
- `GET /api/logs?source=database` - Filter logs by source
- `GET /api/logs/{id}` - Get a specific log by ID
- `POST /api/logs` - Create a new log
- `PUT /api/logs/{id}` - Update a log
- `DELETE /api/logs/{id}` - Delete a log

### Users

- `GET /api/users` - Get all users
- `GET /api/users/{username}` - Get a specific user by username
- `POST /api/users` - Create a new user
- `PUT /api/users/{username}` - Update a user
- `DELETE /api/users/{username}` - Delete a user

### Metrics

- `GET /api/metrics` - Get all metrics

### Config

- `GET /api/config` - Get current configuration
- `POST /api/config` - Update configuration

## Web UI

The web UI is available at:

- `/dashboard` - Main dashboard with stats and recent logs
- `/` - Redirects to dashboard

## Example API Requests

### Create a Log Entry

```bash
curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer fake-token-for-testing" \
  -d '{
    "level": "error",
    "message": "Failed to connect to database",
    "source": "database",
    "data": {"error_code": 1045, "retries": 3}
  }'
```

### Create a User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer fake-token-for-testing" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "roles": ["user"]
  }'
```

## Refactoring Ideas

Here are some ideas for refactoring this code:

1. Introduce proper middleware for authentication and logging
2. Use a router package (gorilla/mux, chi, etc.)
3. Separate handlers into different files by resource
4. Create a proper data layer (repositories)
5. Use interfaces for dependency injection
6. Move validation to separate package
7. Implement proper error handling
8. Use contexts for cancellation
9. Separate frontend from backend
10. Use proper templating system
11. Add tests
12. Add configuration using environment variables
13. Implement logging
14. Use proper filtering with query parameters
15. Implement pagination
