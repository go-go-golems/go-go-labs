# Stream Task Overview API Documentation

## Overview

This API provides endpoints to manage live streaming task information with features for tracking stream metadata and task steps. The API follows REST principles and returns data in JSON format.

## Base URL

All endpoints are relative to the base URL: `http://localhost:8080`

## Authentication

Currently, the API does not require authentication.

## Data Models

### StreamInfo

Represents metadata about the stream.

```json
{
  "title": "Building a React Component Library",
  "description": "Creating reusable UI components with TailwindCSS",
  "startTime": "2025-05-11T14:30:45Z",
  "language": "JavaScript/React",
  "githubRepo": "https://github.com/yourusername/component-library",
  "viewerCount": 42
}
```

### StepInfo

Represents the task steps for the stream.

```json
{
  "completed": ["Project setup and initialization", "Design system planning"],
  "active": "Setting up component architecture",
  "upcoming": ["Implement Button component", "Create Card component"]
}
```

## API Endpoints

### Stream Information

#### Get Stream Information

Returns current stream information.

- **URL**: `/api/stream`
- **Method**: `GET`
- **Response**: `StreamInfo` object

**Example Response:**

```json
{
  "title": "Building a React Component Library",
  "description": "Creating reusable UI components with TailwindCSS",
  "startTime": "2025-05-11T14:30:45Z",
  "language": "JavaScript/React",
  "githubRepo": "https://github.com/yourusername/component-library",
  "viewerCount": 42
}
```

#### Update Stream Information

Updates stream information.

- **URL**: `/api/stream`
- **Method**: `PUT`
- **Content-Type**: `application/json`
- **Request Body**: `StreamInfo` object
- **Response**: Updated `StreamInfo` object

**Example Request:**

```json
{
  "title": "Building a Component Library in React",
  "description": "Creating a design system with TailwindCSS",
  "startTime": "2025-05-11T14:30:45Z",
  "language": "JavaScript/React",
  "githubRepo": "https://github.com/yourusername/component-library",
  "viewerCount": 45
}
```

### Task Steps

#### Get All Steps

Returns all task steps (completed, active, and upcoming).

- **URL**: `/api/stream/steps`
- **Method**: `GET`
- **Response**: `StepInfo` object

**Example Response:**

```json
{
  "completed": ["Project setup and initialization", "Design system planning"],
  "active": "Setting up component architecture",
  "upcoming": ["Implement Button component", "Create Card component", "Build Form elements"]
}
```

#### Set Active Step

Sets a new active step. The current active step is automatically moved to completed.

- **URL**: `/api/stream/steps/active`
- **Method**: `PUT`
- **Content-Type**: `application/json`
- **Request Body**: 
  ```json
  {
    "step": "New active step name"
  }
  ```
- **Response**: Updated `StepInfo` object

**Example Request:**

```json
{
  "step": "Implementing component architecture"
}
```

#### Add Upcoming Step

Adds a new step to the upcoming list.

- **URL**: `/api/stream/steps/upcoming`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body**: 
  ```json
  {
    "step": "New step name"
  }
  ```
- **Response**: Updated `StepInfo` object

**Example Request:**

```json
{
  "step": "Add animation effects"
}
```

#### Complete Active Step

Completes the current active step and makes the next upcoming step active.

- **URL**: `/api/stream/steps/complete`
- **Method**: `POST`
- **Response**: Updated `StepInfo` object

#### Reactivate Step

Moves a step from either completed or upcoming lists to active.

- **URL**: `/api/stream/steps/reactivate`
- **Method**: `PUT`
- **Content-Type**: `application/json`
- **Request Body**: 
  ```json
  {
    "step": "Step to reactivate",
    "source": "completed or upcoming"
  }
  ```
- **Response**: Updated `StepInfo` object

**Example Request:**

```json
{
  "step": "Design system planning",
  "source": "completed"
}
```

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200 OK`: Request successful
- `400 Bad Request`: Invalid request data or parameters
- `500 Internal Server Error`: Server-side error

Errors are returned as JSON objects:

```json
{
  "error": "Error message"
}
```

## Database Schema

The backend uses SQLite with two tables:

### stream_info

| Column      | Type     | Description           |
|-------------|---------|-----------------------|
| id          | INTEGER  | Primary key (always 1)|
| title       | TEXT     | Stream title          |
| description | TEXT     | Stream description    |
| start_time  | DATETIME | Stream start time     |
| language    | TEXT     | Programming language  |
| github_repo | TEXT     | GitHub repository URL |
| viewer_count| INTEGER  | Number of viewers     |

### steps

| Column    | Type  | Description                 |
|-----------|-------|-----------------------------|  
| id        | INTEGER | Primary key (always 1)     |
| completed | TEXT    | JSON array of completed steps |
| active    | TEXT    | Current active step        |
| upcoming  | TEXT    | JSON array of upcoming steps  |