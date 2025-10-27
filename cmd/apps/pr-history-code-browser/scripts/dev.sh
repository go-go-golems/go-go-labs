#!/bin/bash
# Development script that runs both backend and frontend

set -e

DB_PATH="${DB_PATH:-/home/manuel/workspaces/2025-10-16/add-gpt5-responses-to-geppetto/geppetto/ttmp/2025-10-23/git-history-and-code-index.db}"
PORT="${PORT:-8080}"

echo "Starting PR History & Code Browser in development mode..."
echo "Database: $DB_PATH"
echo "Backend Port: $PORT"
echo "Frontend: http://localhost:5173"
echo ""

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database not found at $DB_PATH"
    echo "Please set DB_PATH environment variable or update the script"
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Shutting down..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    exit 0
}

trap cleanup EXIT INT TERM

# Start backend
echo "Starting Go backend..."
go run main.go --db "$DB_PATH" --dev --port "$PORT" &
BACKEND_PID=$!

# Wait a bit for backend to start
sleep 2

# Start frontend
echo "Starting Vite frontend..."
cd frontend
npm run dev &
FRONTEND_PID=$!

# Wait for both processes
wait

