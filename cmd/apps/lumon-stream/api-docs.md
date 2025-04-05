# API Documentation

This document provides detailed information about the API endpoints available in the LumonStream application.

## Base URL

All API endpoints are relative to the base URL: `http://localhost:8080/api`

## Authentication

Currently, the API does not require authentication.

## Endpoints

### Stream Information

#### Get Stream Information

Retrieves the current stream information, including completed, active, and upcoming steps.

- **URL**: `/stream-info`
- **Method**: `GET`
- **Response Format**: JSON

**Success Response:**
```json
{
  "success": true,
  "data": {
    "StreamInfo": {
      "id": 1,
      "title": "Building a React Component Library",
      "description": "Creating reusable UI components with TailwindCSS",
      "startTime": "2025-04-05T19:09:31Z",
      "language": "JavaScript/React",
      "githubRepo": "https://github.com/yourusername/component-library",
      "viewerCount": 42
    },
    "CompletedSteps": [
      {
        "id": 1,
        "content": "Project setup and initialization",
        "status": "completed",
        "createdAt": "2025-04-05T19:09:31Z"
      },
      {
        "id": 2,
        "content": "Design system planning",
        "status": "completed",
        "createdAt": "2025-04-05T19:09:31Z"
      }
    ],
    "ActiveStep": {
      "id": 3,
      "content": "Setting up component architecture",
      "status": "active",
      "createdAt": "2025-04-05T19:09:31Z"
    },
    "UpcomingSteps": [
      {
        "id": 4,
        "content": "Implement Button component",
        "status": "upcoming",
        "createdAt": "2025-04-05T19:09:31Z"
      },
      {
        "id": 5,
        "content": "Create Card component",
        "status": "upcoming",
        "createdAt": "2025-04-05T19:09:31Z"
      }
    ]
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Failed to retrieve stream information"
}
```

#### Update Stream Information

Updates the stream information.

- **URL**: `/stream-info`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body**:
```json
{
  "id": 1,
  "title": "Updated Stream Title",
  "description": "Updated stream description",
  "startTime": "2025-04-05T19:09:31Z",
  "language": "TypeScript/React",
  "githubRepo": "https://github.com/yourusername/updated-repo",
  "viewerCount": 100
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Stream information updated successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Failed to update stream information"
}
```

### Steps Management

#### Add Step

Adds a new step to the stream.

- **URL**: `/steps`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body**:
```json
{
  "content": "New step content",
  "status": "upcoming"
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Step added successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Failed to add step"
}
```

#### Update Step Status

Updates the status of an existing step.

- **URL**: `/steps/status`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body**:
```json
{
  "id": 4,
  "status": "active"
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Step status updated successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Failed to update step status"
}
```

## Status Codes

- **200 OK**: The request was successful
- **400 Bad Request**: The request was invalid or cannot be served
- **500 Internal Server Error**: An error occurred on the server

## Data Models

### StreamInfo

| Field       | Type     | Description                           |
|-------------|----------|---------------------------------------|
| id          | integer  | Unique identifier                     |
| title       | string   | Stream title                          |
| description | string   | Stream description                    |
| startTime   | datetime | Stream start time                     |
| language    | string   | Programming language or framework     |
| githubRepo  | string   | GitHub repository URL                 |
| viewerCount | integer  | Current viewer count                  |

### Step

| Field     | Type     | Description                                      |
|-----------|----------|--------------------------------------------------|
| id        | integer  | Unique identifier                                |
| content   | string   | Step content                                     |
| status    | string   | Step status (completed, active, or upcoming)     |
| createdAt | datetime | Creation timestamp                               |
