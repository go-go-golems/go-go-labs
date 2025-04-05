# Building a Modern Web Application with Go, React, RTK-Query, and Bun

This comprehensive tutorial will guide you through building a complete web application with a Golang backend, SQLite database, React frontend with RTK-Query, and a CLI tool. We'll also cover how to use Bun for React compilation and embed the built files in the Golang binary.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Prerequisites](#prerequisites)
3. [Project Structure](#project-structure)
4. [Backend Implementation](#backend-implementation)
   - [Setting Up Go Modules](#setting-up-go-modules)
   - [Implementing SQLite Database](#implementing-sqlite-database)
   - [Creating API Endpoints](#creating-api-endpoints)
   - [Handling CORS](#handling-cors)
5. [Frontend Implementation](#frontend-implementation)
   - [Setting Up React with RTK-Query](#setting-up-react-with-rtk-query)
   - [Implementing the Stream Info Display Component](#implementing-the-stream-info-display-component)
   - [Connecting to the Backend API](#connecting-to-the-backend-api)
6. [CLI Implementation](#cli-implementation)
   - [Setting Up Cobra](#setting-up-cobra)
   - [Implementing Commands](#implementing-commands)
7. [Build System](#build-system)
   - [Configuring Bun for React Compilation](#configuring-bun-for-react-compilation)
   - [Embedding Built Files in Golang Binary](#embedding-built-files-in-golang-binary)
   - [Debug vs. Production Mode](#debug-vs-production-mode)
8. [Running the Application](#running-the-application)
   - [Development Mode](#development-mode)
   - [Production Mode](#production-mode)
9. [Conclusion](#conclusion)

## Project Overview

Our project, "LumonStream," is a web application for live coding streamers to display information about their stream, track progress, and manage tasks. The application features:

- A Golang backend with SQLite database
- A React frontend with RTK-Query for state management
- A CLI tool for interacting with the server
- A build system using Bun for React compilation
- File embedding to bundle the frontend with the backend binary

The UI design is inspired by the TV show "Severance," with a minimalist corporate aesthetic.

## Prerequisites

To follow this tutorial, you'll need:

- Go 1.18 or later
- Node.js 16 or later
- Bun (for React compilation)
- Basic knowledge of Go, React, and JavaScript

## Project Structure

Our project has the following structure:

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
└── package.json
```

Let's start by implementing each component of the application.

## Backend Implementation

### Setting Up Go Modules

First, let's set up the Go modules for our backend:

```bash
mkdir -p LumonStream/backend
cd LumonStream/backend
go mod init github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend
```

Next, we'll install the necessary dependencies:

```bash
go get github.com/mattn/go-sqlite3 github.com/gorilla/mux github.com/rs/cors
```

### Implementing SQLite Database

Let's create the data models for our application. Create a file at `backend/models/stream_info.go`:

```go
package models

import (
	"time"
)

// StreamInfo represents the stream information data structure
type StreamInfo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	Language    string    `json:"language"`
	GithubRepo  string    `json:"githubRepo"`
	ViewerCount int       `json:"viewerCount"`
}

// Step represents a task step in the stream
type Step struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"` // "completed", "active", or "upcoming"
	CreatedAt time.Time `json:"createdAt"`
}
```

Now, let's implement the database layer at `backend/database/database.go`:

```go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/models"
)

// DB is the database connection
var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return err
	}

	// Initialize with default data if empty
	err = initializeDefaultData()
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the necessary tables in the database
func createTables() error {
	streamInfoTable := `
	CREATE TABLE IF NOT EXISTS stream_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		start_time DATETIME NOT NULL,
		language TEXT,
		github_repo TEXT,
		viewer_count INTEGER DEFAULT 0
	);`

	stepsTable := `
	CREATE TABLE IF NOT EXISTS steps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	_, err := DB.Exec(streamInfoTable)
	if err != nil {
		return fmt.Errorf("error creating stream_info table: %w", err)
	}

	_, err = DB.Exec(stepsTable)
	if err != nil {
		return fmt.Errorf("error creating steps table: %w", err)
	}

	return nil
}

// initializeDefaultData adds default data if the database is empty
func initializeDefaultData() error {
	// Check if stream_info table is empty
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM stream_info").Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking stream_info table: %w", err)
	}

	if count == 0 {
		// Insert default stream info
		_, err = DB.Exec(`
			INSERT INTO stream_info (title, description, start_time, language, github_repo, viewer_count)
			VALUES (?, ?, ?, ?, ?, ?)`,
			"Building a React Component Library",
			"Creating reusable UI components with TailwindCSS",
			time.Now(),
			"JavaScript/React",
			"https://github.com/yourusername/component-library",
			42,
		)
		if err != nil {
			return fmt.Errorf("error inserting default stream info: %w", err)
		}

		// Insert default steps
		completedSteps := []string{
			"Project setup and initialization",
			"Design system planning",
		}
		for _, step := range completedSteps {
			_, err = DB.Exec(`
				INSERT INTO steps (content, status, created_at)
				VALUES (?, ?, ?)`,
				step,
				"completed",
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("error inserting completed step: %w", err)
			}
		}

		// Insert active step
		_, err = DB.Exec(`
			INSERT INTO steps (content, status, created_at)
			VALUES (?, ?, ?)`,
			"Setting up component architecture",
			"active",
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("error inserting active step: %w", err)
		}

		// Insert upcoming steps
		upcomingSteps := []string{
			"Implement Button component",
			"Create Card component",
			"Build Form elements",
			"Add dark mode toggle",
		}
		for _, step := range upcomingSteps {
			_, err = DB.Exec(`
				INSERT INTO steps (content, status, created_at)
				VALUES (?, ?, ?)`,
				step,
				"upcoming",
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("error inserting upcoming step: %w", err)
			}
		}
	}

	return nil
}

// GetStreamInfo retrieves the stream information from the database
func GetStreamInfo() (models.StreamInfo, error) {
	var info models.StreamInfo
	err := DB.QueryRow(`
		SELECT id, title, description, start_time, language, github_repo, viewer_count
		FROM stream_info
		ORDER BY id DESC
		LIMIT 1
	`).Scan(
		&info.ID,
		&info.Title,
		&info.Description,
		&info.StartTime,
		&info.Language,
		&info.GithubRepo,
		&info.ViewerCount,
	)
	if err != nil {
		return models.StreamInfo{}, fmt.Errorf("error getting stream info: %w", err)
	}
	return info, nil
}

// UpdateStreamInfo updates the stream information in the database
func UpdateStreamInfo(info models.StreamInfo) error {
	_, err := DB.Exec(`
		UPDATE stream_info
		SET title = ?, description = ?, start_time = ?, language = ?, github_repo = ?, viewer_count = ?
		WHERE id = ?
	`,
		info.Title,
		info.Description,
		info.StartTime,
		info.Language,
		info.GithubRepo,
		info.ViewerCount,
		info.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating stream info: %w", err)
	}
	return nil
}

// GetSteps retrieves all steps from the database grouped by status
func GetSteps() ([]models.Step, []models.Step, []models.Step, error) {
	var completedSteps, activeSteps, upcomingSteps []models.Step

	// Get completed steps
	rows, err := DB.Query(`
		SELECT id, content, status, created_at
		FROM steps
		WHERE status = 'completed'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting completed steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step models.Step
		err := rows.Scan(&step.ID, &step.Content, &step.Status, &step.CreatedAt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error scanning completed step: %w", err)
		}
		completedSteps = append(completedSteps, step)
	}

	// Get active step
	rows, err = DB.Query(`
		SELECT id, content, status, created_at
		FROM steps
		WHERE status = 'active'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting active step: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step models.Step
		err := rows.Scan(&step.ID, &step.Content, &step.Status, &step.CreatedAt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error scanning active step: %w", err)
		}
		activeSteps = append(activeSteps, step)
	}

	// Get upcoming steps
	rows, err = DB.Query(`
		SELECT id, content, status, created_at
		FROM steps
		WHERE status = 'upcoming'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting upcoming steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step models.Step
		err := rows.Scan(&step.ID, &step.Content, &step.Status, &step.CreatedAt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error scanning upcoming step: %w", err)
		}
		upcomingSteps = append(upcomingSteps, step)
	}

	return completedSteps, activeSteps, upcomingSteps, nil
}

// UpdateStepStatus updates the status of a step
func UpdateStepStatus(id int, status string) error {
	_, err := DB.Exec(`
		UPDATE steps
		SET status = ?
		WHERE id = ?
	`, status, id)
	if err != nil {
		return fmt.Errorf("error updating step status: %w", err)
	}
	return nil
}

// AddStep adds a new step to the database
func AddStep(content, status string) error {
	_, err := DB.Exec(`
		INSERT INTO steps (content, status, created_at)
		VALUES (?, ?, ?)
	`, content, status, time.Now())
	if err != nil {
		return fmt.Errorf("error adding step: %w", err)
	}
	return nil
}
```

### Creating API Endpoints

Now, let's implement the API handlers at `backend/handlers/handlers.go`:

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/database"
	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/models"
)

// Response is a generic API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// StreamInfoResponse combines stream info and steps
type StreamInfoResponse struct {
	models.StreamInfo
	CompletedSteps []models.Step `json:"completedSteps"`
	ActiveStep     *models.Step  `json:"activeStep"`
	UpcomingSteps  []models.Step `json:"upcomingSteps"`
}

// GetStreamInfo handles GET requests for stream information
func GetStreamInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get stream info
	info, err := database.GetStreamInfo()
	if err != nil {
		log.Printf("Error getting stream info: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to retrieve stream information",
		})
		return
	}

	// Get steps
	completedSteps, activeSteps, upcomingSteps, err := database.GetSteps()
	if err != nil {
		log.Printf("Error getting steps: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to retrieve steps",
		})
		return
	}

	// Prepare response
	response := StreamInfoResponse{
		StreamInfo:     info,
		CompletedSteps: completedSteps,
		UpcomingSteps:  upcomingSteps,
	}

	// Set active step if available
	if len(activeSteps) > 0 {
		response.ActiveStep = &activeSteps[0]
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    response,
	})
}

// UpdateStreamInfo handles POST requests to update stream information
func UpdateStreamInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var info models.StreamInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if info.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Title is required",
		})
		return
	}

	// Ensure StartTime is set
	if info.StartTime.IsZero() {
		info.StartTime = time.Now()
	}

	// Update stream info
	err = database.UpdateStreamInfo(info)
	if err != nil {
		log.Printf("Error updating stream info: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to update stream information",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Stream information updated successfully",
	})
}

// AddStep handles POST requests to add a new step
func AddStep(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var step struct {
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&step)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if step.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Step content is required",
		})
		return
	}

	// Validate status
	if step.Status != "completed" && step.Status != "active" && step.Status != "upcoming" {
		step.Status = "upcoming" // Default to upcoming if invalid
	}

	// Add step
	err = database.AddStep(step.Content, step.Status)
	if err != nil {
		log.Printf("Error adding step: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to add step",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Step added successfully",
	})
}

// UpdateStepStatus handles POST requests to update a step's status
func UpdateStepStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var request struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if request.ID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Valid step ID is required",
		})
		return
	}

	// Validate status
	if request.Status != "completed" && request.Status != "active" && request.Status != "upcoming" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Valid status is required (completed, active, or upcoming)",
		})
		return
	}

	// Update step status
	err = database.UpdateStepStatus(request.ID, request.Status)
	if err != nil {
		log.Printf("Error updating step status: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to update step status",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Step status updated successfully",
	})
}
```

### Handling CORS

Let's implement the main server file at `backend/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/database"
	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/handlers"
	"github.com/rs/cors"
)

var (
	port      int
	dbPath    string
	debugMode bool
)

func init() {
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.StringVar(&dbPath, "db", "./lumonstream.db", "Path to SQLite database file")
	flag.BoolVar(&debugMode, "debug", false, "Run in debug mode")
}

func main() {
	flag.Parse()

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	// Initialize database
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/stream-info", handlers.GetStreamInfo).Methods("GET")
	api.HandleFunc("/stream-info", handlers.UpdateStreamInfo).Methods("POST")
	api.HandleFunc("/steps", handlers.AddStep).Methods("POST")
	api.HandleFunc("/steps/status", handlers.UpdateStepStatus).Methods("POST")

	// Serve static files in production mode
	if !debugMode {
		// When built with -tags embed, this will use the embedded files
		// Otherwise, this function is a no-op in debug mode
		SetupStaticFiles(r)
	}

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, you should restrict this
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Create server
	handler := c.Handler(r)

	// Start server
	serverAddr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Starting server in %s mode on %s", map[bool]string{true: "debug", false: "production"}[debugMode], serverAddr)
	log.Printf("Database path: %s", dbPath)
	log.Fatal(http.ListenAndServe(serverAddr, handler))
}
```

## Frontend Implementation

### Setting Up React with RTK-Query

First, let's set up the React frontend:

```bash
mkdir -p LumonStream/frontend
cd LumonStream/frontend
npx create-react-app .
npm install @reduxjs/toolkit react-redux
npm install tailwindcss postcss autoprefixer lucide-react
```

Now, let's create the Redux store at `frontend/src/app/store.js`:

```javascript
import { configureStore } from '@reduxjs/toolkit';
import { apiSlice } from '../features/api/apiSlice';

export const store = configureStore({
  reducer: {
    [apiSlice.reducerPath]: apiSlice.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(apiSlice.middleware),
  devTools: process.env.NODE_ENV !== 'production',
});
```

Let's implement the API slice at `frontend/src/features/api/apiSlice.js`:

```javascript
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

// Define our single API slice object
export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: 'http://localhost:8080/api' }),
  tagTypes: ['StreamInfo'],
  endpoints: (builder) => ({
    getStreamInfo: builder.query({
      query: () => '/stream-info',
      transformResponse: (response) => {
        // Extract the data from the response
        if (response.success && response.data) {
          return {
            ...response.data.StreamInfo,
            completedSteps: response.data.CompletedSteps.map(step => step.content),
            activeStep: response.data.ActiveStep ? response.data.ActiveStep.content : "",
            upcomingSteps: response.data.UpcomingSteps.map(step => step.content),
          };
        }
        return null;
      },
      providesTags: ['StreamInfo'],
    }),
    updateStreamInfo: builder.mutation({
      query: (streamInfo) => ({
        url: '/stream-info',
        method: 'POST',
        body: streamInfo,
      }),
      invalidatesTags: ['StreamInfo'],
    }),
    addStep: builder.mutation({
      query: (step) => ({
        url: '/steps',
        method: 'POST',
        body: step,
      }),
      invalidatesTags: ['StreamInfo'],
    }),
    updateStepStatus: builder.mutation({
      query: (data) => ({
        url: '/steps/status',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['StreamInfo'],
    }),
  }),
});

// Export the auto-generated hooks for the endpoints
export const { 
  useGetStreamInfoQuery, 
  useUpdateStreamInfoMutation,
  useAddStepMutation,
  useUpdateStepStatusMutation
} = apiSlice;
```

Update the `frontend/src/index.js` file to use the Redux store:

```javascript
import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import { store } from './app/store';
import { Provider } from 'react-redux';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>
);
```

### Implementing the Stream Info Display Component

Let's create the StreamInfoDisplay component at `frontend/src/components/StreamInfoDisplay.jsx`:

```jsx
import React, { useState, useEffect } from 'react';
import { Clock, Code, MessagesSquare, Users, Info, Github, Check, List, ArrowRight } from 'lucide-react';
import { useGetStreamInfoQuery, useUpdateStreamInfoMutation, useAddStepMutation, useUpdateStepStatusMutation } from '../features/api/apiSlice';

const StreamInfoDisplay = () => {
  // Use RTK Query to fetch stream information
  const { data: apiStreamInfo, isLoading, isError, refetch } = useGetStreamInfoQuery(null, {
    pollingInterval: 10000, // Poll every 10 seconds for updates
  });
  
  const [updateStreamInfo] = useUpdateStreamInfoMutation();
  const [addStep] = useAddStepMutation();
  const [updateStepStatus] = useUpdateStepStatusMutation();

  // State to store all the stream information
  const [streamInfo, setStreamInfo] = useState({
    title: "Building a React Component Library",
    description: "Creating reusable UI components with TailwindCSS",
    startTime: new Date().toISOString(),
    language: "JavaScript/React",
    githubRepo: "https://github.com/yourusername/component-library",
    viewerCount: 42,
  });

  // Update local state when API data is received
  useEffect(() => {
    if (apiStreamInfo) {
      setStreamInfo(apiStreamInfo);
      setEditableInfo(apiStreamInfo);
      
      // Update steps from API
      if (apiStreamInfo.completedSteps) {
        setCompletedSteps(apiStreamInfo.completedSteps);
      }
      
      if (apiStreamInfo.activeStep) {
        setActiveStep(apiStreamInfo.activeStep);
      }
      
      if (apiStreamInfo.upcomingSteps) {
        setUpcomingSteps(apiStreamInfo.upcomingSteps);
      }
    }
  }, [apiStreamInfo]);

  // State for steps (completed, current, upcoming)
  const [completedSteps, setCompletedSteps] = useState([
    "Project setup and initialization",
    "Design system planning"
  ]);
  
  const [activeStep, setActiveStep] = useState("Setting up component architecture");
  
  const [upcomingSteps, setUpcomingSteps] = useState([
    "Implement Button component",
    "Create Card component",
    "Build Form elements",
    "Add dark mode toggle"
  ]);

  // State for new step input
  const [newStep, setNewStep] = useState("");
  const [newTopic, setNewTopic] = useState("");

  // State for editing mode
  const [isEditing, setIsEditing] = useState(false);
  const [editableInfo, setEditableInfo] = useState({...streamInfo});
  
  // Calculate stream duration
  const [duration, setDuration] = useState("00:00:00");
  
  useEffect(() => {
    const updateDuration = () => {
      const start = new Date(streamInfo.startTime);
      const now = new Date();
      const diff = Math.floor((now - start) / 1000);
      
      const hours = Math.floor(diff / 3600).toString().padStart(2, '0');
      const minutes = Math.floor((diff % 3600) / 60).toString().padStart(2, '0');
      const seconds = (diff % 60).toString().padStart(2, '0');
      
      setDuration(`${hours}:${minutes}:${seconds}`);
    };
    
    updateDuration();
    const interval = setInterval(updateDuration, 1000);
    
    return () => clearInterval(interval);
  }, [streamInfo.startTime]);
  
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setEditableInfo(prev => ({...prev, [name]: value}));
  };
  
  const saveChanges = async () => {
    setStreamInfo({...editableInfo});
    setIsEditing(false);
    
    // Update the backend via API
    try {
      await updateStreamInfo({...editableInfo});
    } catch (error) {
      console.error("Failed to update stream info:", error);
    }
  };
  
  const cancelChanges = () => {
    setEditableInfo({...streamInfo});
    setIsEditing(false);
  };
  
  const resetTimer = () => {
    const newInfo = {...streamInfo, startTime: new Date().toISOString()};
    setStreamInfo(newInfo);
    setEditableInfo(newInfo);
    
    // Update the backend via API
    try {
      updateStreamInfo(newInfo);
    } catch (error) {
      console.error("Failed to update timer:", error);
    }
  };

  const addNewStep = async () => {
    if (newStep.trim()) {
      // Update local state
      setUpcomingSteps([...upcomingSteps, newStep.trim()]);
      setNewStep("");
      
      // Update backend via API
      try {
        await addStep({
          content: newStep.trim(),
          status: "upcoming"
        });
      } catch (error) {
        console.error("Failed to add step:", error);
      }
    }
  };

  const setNewActiveTopic = async () => {
    if (newTopic.trim()) {
      // Update local state
      if (activeStep) {
        setCompletedSteps([...completedSteps, activeStep]);
      }
      setActiveStep(newTopic.trim());
      setNewTopic("");
      
      // Update backend via API
      try {
        // First add the new active step
        await addStep({
          content: newTopic.trim(),
          status: "active"
        });
        
        // Then refresh to get updated data
        refetch();
      } catch (error) {
        console.error("Failed to set new topic:", error);
      }
    }
  };

  const completeCurrentStep = async () => {
    if (activeStep) {
      // Update local state
      setCompletedSteps([...completedSteps, activeStep]);
      if (upcomingSteps.length > 0) {
        setActiveStep(upcomingSteps[0]);
        setUpcomingSteps(upcomingSteps.slice(1));
      } else {
        setActiveStep("");
      }
      
      // Update backend via API
      try {
        // Mark current step as completed
        await updateStepStatus({
          id: apiStreamInfo.activeStepId,
          status: "completed"
        });
        
        // If there's an upcoming step, make it active
        if (upcomingSteps.length > 0 && apiStreamInfo.upcomingStepIds && apiStreamInfo.upcomingStepIds.length > 0) {
          await updateStepStatus({
            id: apiStreamInfo.upcomingStepIds[0],
            status: "active"
          });
        }
        
        // Refresh to get updated data
        refetch();
      } catch (error) {
        console.error("Failed to complete step:", error);
      }
    }
  };

  const makeStepActive = async (step, source) => {
    // Update local state
    if (activeStep) {
      setCompletedSteps([...completedSteps, activeStep]);
    }
    
    setActiveStep(step);
    
    if (source === 'upcoming') {
      setUpcomingSteps(upcomingSteps.filter(s => s !== step));
    } else if (source === 'completed') {
      setCompletedSteps(completedSteps.filter(s => s !== step));
    }
    
    // Update backend via API
    try {
      // Find the step ID based on the source
      let stepId;
      if (source === 'upcoming' && apiStreamInfo.upcomingStepIds) {
        const index = apiStreamInfo.upcomingSteps.indexOf(step);
        if (index !== -1) {
          stepId = apiStreamInfo.upcomingStepIds[index];
        }
      } else if (source === 'completed' && apiStreamInfo.completedStepIds) {
        const index = apiStreamInfo.completedSteps.indexOf(step);
        if (index !== -1) {
          stepId = apiStreamInfo.completedStepIds[index];
        }
      }
      
      if (stepId) {
        // If there's an active step, mark it as completed or upcoming
        if (apiStreamInfo.activeStepId) {
          await updateStepStatus({
            id: apiStreamInfo.activeStepId,
            status: source === 'completed' ? "completed" : "upcoming"
          });
        }
        
        // Mark the selected step as active
        await updateStepStatus({
          id: stepId,
          status: "active"
        });
        
        // Refresh to get updated data
        refetch();
      }
    } catch (error) {
      console.error("Failed to make step active:", error);
    }
  };

  if (isLoading) {
    return <div className="w-full max-w-4xl mx-auto p-6 bg-white text-black rounded-none shadow-lg font-mono">Loading...</div>;
  }

  if (isError) {
    return <div className="w-full max-w-4xl mx-auto p-6 bg-white text-black rounded-none shadow-lg font-mono">
      Error loading stream information. Please make sure the backend server is running.
    </div>;
  }

  return (
    <div className="w-full max-w-4xl mx-auto p-6 bg-white text-black rounded-none shadow-lg font-mono" style={{fontFamily: 'monospace'}}>
      <div className="border-b-2 border-black pb-4 mb-6">
        <div className="flex justify-between items-center">
          <h1 className="text-2xl font-bold uppercase tracking-widest">LUMON INDUSTRIES</h1>
          <div className="flex items-center">
            <div className="flex flex-col items-end mr-6">
              <div className="text-xs uppercase">MACRODATA STREAM</div>
              <div className="text-xl font-bold">{duration}</div>
            </div>
            {!isEditing ? (
              <button 
                onClick={() => setIsEditing(true)}
                className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
              >
                Edit Parameters
              </button>
            ) : (
              <>
                <button 
                  onClick={saveChanges}
                  className="px-4 py-2 bg-green-900 text-white rounded-none hover:bg-green-800 transition-colors uppercase text-xs tracking-wider mr-2"
                >
                  Save
                </button>
                <button 
                  onClick={cancelChanges}
                  className="px-4 py-2 bg-red-900 text-white rounded-none hover:bg-red-800 transition-colors uppercase text-xs tracking-wider"
                >
                  Cancel
                </button>
              </>
            )}
          </div>
        </div>
      </div>
      
      {isEditing ? (
        <div className="grid grid-cols-1 gap-4 p-4 border-2 border-black">
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Stream Title</label>
            <input
              name="title"
              value={editableInfo.title}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Description</label>
            <textarea
              name="description"
              value={editableInfo.description}
              onChange={handleInputChange}
              rows="2"
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Programming Language/Framework</label>
            <input
              name="language"
              value={editableInfo.language}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">GitHub Repository</label>
            <input
              name="githubRepo"
              value={editableInfo.githubRepo}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium mb-1 uppercase tracking-wider">Viewer Count</label>
            <input
              type="number"
              name="viewerCount"
              value={editableInfo.viewerCount}
              onChange={handleInputChange}
              className="w-full p-2 border-2 border-black rounded-none bg-white text-black"
            />
          </div>
          
          <div className="flex items-center">
            <button
              onClick={resetTimer}
              className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
            >
              Reset Timer
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <div className="lg:col-span-5 border-2 border-black p-4">
            <div className="mb-6">
              <div className="text-xs uppercase tracking-wider mb-1">PROJECT DESIGNATION</div>
              <h2 className="font-bold text-lg">{streamInfo.title}</h2>
              <p className="text-sm mt-1">{streamInfo.description}</p>
            </div>
            
            <div className="flex items-center mb-3">
              <div className="w-6 h-6 mr-2 flex items-center justify-center bg-black text-white">
                <Code size={16} />
              </div>
              <div>
                <div className="text-xs uppercase tracking-wider">LANGUAGE</div>
                <span>{streamInfo.language}</span>
              </div>
            </div>
            
            <div className="flex items-center mb-3">
              <div className="w-6 h-6 mr-2 flex items-center justify-center bg-black text-white">
                <Github size={16} />
              </div>
              <div>
                <div className="text-xs uppercase tracking-wider">REPOSITORY</div>
                <a 
                  href={streamInfo.githubRepo} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-900 hover:underline break-all"
                >
                  {streamInfo.githubRepo}
                </a>
              </div>
            </div>
            
            <div className="flex items-center">
              <div className="w-6 h-6 mr-2 flex items-center justify-center bg-black text-white">
                <Users size={16} />
              </div>
              <div>
                <div className="text-xs uppercase tracking-wider">VIEWERS</div>
                <span>{streamInfo.viewerCount}</span>
              </div>
            </div>
            
            <div className="mt-6">
              <div className="flex">
                <input
                  type="text"
                  value={newTopic}
                  onChange={(e) => setNewTopic(e.target.value)}
                  placeholder="Enter new topic..."
                  className="flex-grow p-2 border-2 border-black rounded-none bg-white text-black"
                />
                <button
                  onClick={setNewActiveTopic}
                  className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
                >
                  Set Topic
                </button>
              </div>
            </div>
          </div>
          
          <div className="lg:col-span-7 border-2 border-black">
            <div className="border-b-2 border-black p-3 bg-black text-white">
              <h3 className="uppercase tracking-wider font-bold">CURRENT PROGRESS</h3>
            </div>
            
            {activeStep && (
              <div className="p-4 border-b-2 border-black">
                <div className="flex justify-between items-start">
                  <div>
                    <div className="text-xs uppercase tracking-wider mb-1">ACTIVE TASK</div>
                    <div className="font-bold">{activeStep}</div>
                  </div>
                  <button
                    onClick={completeCurrentStep}
                    className="flex items-center px-3 py-1 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
                  >
                    <Check size={14} className="mr-1" />
                    Complete
                  </button>
                </div>
              </div>
            )}
            
            <div className="p-4 border-b-2 border-black">
              <div className="text-xs uppercase tracking-wider mb-2">COMPLETED TASKS</div>
              {completedSteps.length > 0 ? (
                <ul>
                  {completedSteps.map((step, index) => (
                    <li key={`completed-${index}`} className="mb-2 flex justify-between items-center">
                      <div className="flex items-center">
                        <div className="w-4 h-4 mr-2 flex items-center justify-center bg-black text-white">
                          <Check size={12} />
                        </div>
                        <span className="text-sm">{step}</span>
                      </div>
                      <button
                        onClick={() => makeStepActive(step, 'completed')}
                        className="text-xs text-blue-900 hover:underline"
                      >
                        Reactivate
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <div className="text-sm text-gray-500 italic">No completed tasks yet</div>
              )}
            </div>
            
            <div className="p-4">
              <div className="text-xs uppercase tracking-wider mb-2">UPCOMING TASKS</div>
              {upcomingSteps.length > 0 ? (
                <ul>
                  {upcomingSteps.map((step, index) => (
                    <li key={`upcoming-${index}`} className="mb-2 flex justify-between items-center">
                      <div className="flex items-center">
                        <div className="w-4 h-4 mr-2 flex items-center justify-center border border-black">
                          <ArrowRight size={12} />
                        </div>
                        <span className="text-sm">{step}</span>
                      </div>
                      <button
                        onClick={() => makeStepActive(step, 'upcoming')}
                        className="text-xs text-blue-900 hover:underline"
                      >
                        Make Active
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <div className="text-sm text-gray-500 italic">No upcoming tasks</div>
              )}
              
              <div className="mt-4 flex">
                <input
                  type="text"
                  value={newStep}
                  onChange={(e) => setNewStep(e.target.value)}
                  placeholder="Add new task..."
                  className="flex-grow p-2 border-2 border-black rounded-none bg-white text-black"
                />
                <button
                  onClick={addNewStep}
                  className="px-4 py-2 bg-black text-white rounded-none hover:bg-gray-800 transition-colors uppercase text-xs tracking-wider"
                >
                  Add
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
      
      <div className="mt-4 text-xs text-right">
        <button 
          onClick={() => refetch()} 
          className="text-blue-900 hover:underline"
        >
          Refresh Data
        </button>
      </div>
    </div>
  );
};

export default StreamInfoDisplay;
```

Update the `frontend/src/App.js` file:

```javascript
import React from 'react';
import './App.css';
import StreamInfoDisplay from './components/StreamInfoDisplay';

function App() {
  return (
    <div className="App">
      <StreamInfoDisplay />
    </div>
  );
}

export default App;
```

### Connecting to the Backend API

Update the `frontend/package.json` file to include a proxy for development:

```json
{
  "name": "lumonstream-frontend",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "build:bun": "bun bunbuild.js",
    "test": "react-scripts test",
    "eject": "react-scripts eject"
  },
  "proxy": "http://localhost:8080",
  "eslintConfig": {
    "extends": [
      "react-app",
      "react-app/jest"
    ]
  },
  "browserslist": {
    "production": [
      ">0.2%",
      "not dead",
      "not op_mini all"
    ],
    "development": [
      "last 1 chrome version",
      "last 1 firefox version",
      "last 1 safari version"
    ]
  }
}
```

## CLI Implementation

### Setting Up Cobra

First, let's set up the CLI with Cobra:

```bash
mkdir -p LumonStream/cli
cd LumonStream/cli
go mod init github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/cli
go get github.com/spf13/cobra
```

### Implementing Commands

Let's create the root command at `cli/cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	port      int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lumonstream",
	Short: "LumonStream CLI - Interact with the LumonStream server",
	Long: `LumonStream CLI is a command-line interface for interacting with the LumonStream server.
It allows you to manage stream information, tasks, and server settings.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
}
```

Now, let's implement the get command at `cli/cmd/get.go`:

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

// StreamInfo represents the stream information data structure
type StreamInfo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   string    `json:"startTime"`
	Language    string    `json:"language"`
	GithubRepo  string    `json:"githubRepo"`
	ViewerCount int       `json:"viewerCount"`
}

// Response is a generic API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get stream information",
	Long:  `Retrieve the current stream information from the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/api/stream-info", serverURL)
		
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !response.Success {
			fmt.Printf("Error: %s\n", response.Message)
			return
		}
		
		// Pretty print the data
		prettyJSON, err := json.MarshalIndent(response.Data, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting response: %v\n", err)
			return
		}
		
		fmt.Println(string(prettyJSON))
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
```

Let's implement the update command at `cli/cmd/update.go`:

```go
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	title       string
	description string
	language    string
	githubRepo  string
	viewerCount int
	resetTimer  bool
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update stream information",
	Long:  `Update the stream information on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// First get current info to preserve fields not being updated
		url := fmt.Sprintf("%s/api/stream-info", serverURL)
		
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			return
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !response.Success {
			fmt.Printf("Error: %s\n", response.Message)
			return
		}
		
		// Extract stream info from response
		dataJSON, err := json.Marshal(response.Data)
		if err != nil {
			fmt.Printf("Error processing response: %v\n", err)
			return
		}
		
		var streamInfoResponse struct {
			StreamInfo StreamInfo `json:"StreamInfo"`
		}
		
		if err := json.Unmarshal(dataJSON, &streamInfoResponse); err != nil {
			fmt.Printf("Error extracting stream info: %v\n", err)
			return
		}
		
		info := streamInfoResponse.StreamInfo
		
		// Update fields that were specified
		if cmd.Flags().Changed("title") {
			info.Title = title
		}
		if cmd.Flags().Changed("description") {
			info.Description = description
		}
		if cmd.Flags().Changed("language") {
			info.Language = language
		}
		if cmd.Flags().Changed("github-repo") {
			info.GithubRepo = githubRepo
		}
		if cmd.Flags().Changed("viewer-count") {
			info.ViewerCount = viewerCount
		}
		if resetTimer {
			info.StartTime = time.Now().Format(time.RFC3339)
		}
		
		// Send update request
		jsonData, err := json.Marshal(info)
		if err != nil {
			fmt.Printf("Error preparing request: %v\n", err)
			return
		}
		
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var updateResponse Response
		if err := json.Unmarshal(body, &updateResponse); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !updateResponse.Success {
			fmt.Printf("Error: %s\n", updateResponse.Message)
			return
		}
		
		fmt.Println("Stream information updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	
	// Add flags for all updatable fields
	updateCmd.Flags().StringVar(&title, "title", "", "Stream title")
	updateCmd.Flags().StringVar(&description, "description", "", "Stream description")
	updateCmd.Flags().StringVar(&language, "language", "", "Programming language/framework")
	updateCmd.Flags().StringVar(&githubRepo, "github-repo", "", "GitHub repository URL")
	updateCmd.Flags().IntVar(&viewerCount, "viewer-count", 0, "Current viewer count")
	updateCmd.Flags().BoolVar(&resetTimer, "reset-timer", false, "Reset the stream timer")
}
```

Let's implement the task command at `cli/cmd/task.go`:

```go
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	content string
	status  string
	stepID  int
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage stream tasks",
	Long:  `Add, update, or manage tasks for the stream.`,
}

// addTaskCmd represents the add task command
var addTaskCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task",
	Long:  `Add a new task to the stream.`,
	Run: func(cmd *cobra.Command, args []string) {
		if content == "" {
			fmt.Println("Error: Task content is required")
			return
		}

		// Validate status
		if status != "completed" && status != "active" && status != "upcoming" {
			status = "upcoming" // Default to upcoming if invalid
		}

		url := fmt.Sprintf("%s/api/steps", serverURL)
		
		// Prepare request data
		requestData := map[string]string{
			"content": content,
			"status":  status,
		}
		
		jsonData, err := json.Marshal(requestData)
		if err != nil {
			fmt.Printf("Error preparing request: %v\n", err)
			return
		}
		
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !response.Success {
			fmt.Printf("Error: %s\n", response.Message)
			return
		}
		
		fmt.Println("Task added successfully")
	},
}

// updateTaskCmd represents the update task status command
var updateTaskCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a task's status",
	Long:  `Update the status of an existing task.`,
	Run: func(cmd *cobra.Command, args []string) {
		if stepID <= 0 {
			fmt.Println("Error: Valid step ID is required")
			return
		}

		// Validate status
		if status != "completed" && status != "active" && status != "upcoming" {
			fmt.Println("Error: Valid status is required (completed, active, or upcoming)")
			return
		}

		url := fmt.Sprintf("%s/api/steps/status", serverURL)
		
		// Prepare request data
		requestData := map[string]interface{}{
			"id":     stepID,
			"status": status,
		}
		
		jsonData, err := json.Marshal(requestData)
		if err != nil {
			fmt.Printf("Error preparing request: %v\n", err)
			return
		}
		
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}
		
		if !response.Success {
			fmt.Printf("Error: %s\n", response.Message)
			return
		}
		
		fmt.Println("Task status updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(addTaskCmd)
	taskCmd.AddCommand(updateTaskCmd)
	
	// Add flags for task commands
	addTaskCmd.Flags().StringVar(&content, "content", "", "Task content")
	addTaskCmd.Flags().StringVar(&status, "status", "upcoming", "Task status (completed, active, or upcoming)")
	addTaskCmd.MarkFlagRequired("content")
	
	updateTaskCmd.Flags().IntVar(&stepID, "id", 0, "Task ID")
	updateTaskCmd.Flags().StringVar(&status, "status", "", "Task status (completed, active, or upcoming)")
	updateTaskCmd.MarkFlagRequired("id")
	updateTaskCmd.MarkFlagRequired("status")
}
```

Let's implement the server command at `cli/cmd/server.go`:

```go
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	serverPort int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Control the LumonStream server",
	Long:  `Start, stop, or check the status of the LumonStream server.`,
}

// statusCmd represents the server status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	Long:  `Check if the LumonStream server is running.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/api/stream-info", serverURL)
		
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Server is not running or not accessible at %s\n", serverURL)
			return
		}
		defer resp.Body.Close()
		
		fmt.Printf("Server is running at %s\n", serverURL)
	},
}

// startCmd represents the server start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the LumonStream server on the specified port.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("To start the server, run the following command in a separate terminal:\n\n")
		fmt.Printf("cd /path/to/LumonStream/backend && ./lumonstream --port %d\n\n", serverPort)
		fmt.Printf("Or if you're using the source code directly:\n\n")
		fmt.Printf("cd /path/to/LumonStream/backend && go run main.go --port %d\n", serverPort)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(statusCmd)
	serverCmd.AddCommand(startCmd)
	
	// Add flags for server commands
	startCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
}
```

Finally, let's create the main.go file at `cli/main.go`:

```go
package main

import "github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/cli/cmd"

func main() {
	cmd.Execute()
}
```

## Build System

### Configuring Bun for React Compilation

Let's create a root package.json file at `LumonStream/package.json`:

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

Let's create a Bun build configuration at `frontend/bunbuild.js`:

```javascript
// This file configures the Bun build system for the React frontend
// It will be used to build the React app for production and development

const { build } = require("bun");

async function buildReact() {
  // Build the React app for production
  await build({
    entrypoints: ["./src/index.js"],
    outdir: "./build",
    minify: true,
    target: "browser",
    sourcemap: "external",
  });
  
  console.log("React build completed successfully!");
}

// Export the build function for use in scripts
module.exports = { buildReact };
```

Let's create a development server at `frontend/devserver.js`:

```javascript
// This file configures the development server for the React frontend
// It will be used to serve the React app in development mode

const { serve } = require("bun");

async function startDevServer() {
  const server = serve({
    port: 3000,
    fetch(req) {
      // Forward API requests to the backend
      const url = new URL(req.url);
      if (url.pathname.startsWith('/api')) {
        // Proxy to backend
        return fetch(`http://localhost:8080${url.pathname}${url.search}`, {
          method: req.method,
          headers: req.headers,
          body: req.body
        });
      }
      
      // Serve static files from the public directory
      return new Response("Development server running");
    },
  });
  
  console.log(`Development server running at http://localhost:${server.port}`);
}

// Export the server function for use in scripts
module.exports = { startDevServer };
```

### Embedding Built Files in Golang Binary

Let's create the embed.go file at `backend/embed.go`:

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

Let's create the embed_debug.go file at `backend/embed_debug.go`:

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

Let's create a build script at `backend/build.sh`:

```bash
#!/bin/bash

# This script builds the React frontend and embeds it in the Golang binary

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

### Debug vs. Production Mode

In debug mode, the backend server doesn't serve static files, and the frontend development server handles this. In production mode, the backend server serves the embedded static files.

## Running the Application

### Development Mode

To run the application in development mode:

1. Start the backend server:

```bash
cd LumonStream/backend
go run main.go --debug
```

2. Start the frontend development server:

```bash
cd LumonStream/frontend
npm start
```

3. Access the application at http://localhost:3000

### Production Mode

To run the application in production mode:

1. Build the application:

```bash
cd LumonStream
npm run build:embed
```

2. Run the binary:

```bash
cd LumonStream/backend
./lumonstream
```

3. Access the application at http://localhost:8080

## Conclusion

In this tutorial, we've built a complete web application with a Golang backend, SQLite database, React frontend with RTK-Query, and a CLI tool. We've also set up a build system using Bun for React compilation and implemented file embedding to bundle the frontend with the backend binary.

This architecture provides several benefits:

1. **Single Binary Deployment**: The entire application can be deployed as a single binary, making it easy to distribute and run.
2. **No External Dependencies**: The application doesn't require any external dependencies like Node.js or a separate web server.
3. **Efficient Development Workflow**: The development mode allows for fast iteration with hot reloading.
4. **CLI for Automation**: The CLI tool enables automation and scripting of application interactions.

You can extend this application in various ways:

- Add authentication and user management
- Implement WebSocket for real-time updates
- Add more features to the CLI tool
- Create a mobile app using the same API
- Implement a plugin system for extensibility

Happy coding!
