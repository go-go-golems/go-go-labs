.PHONY: help dev build clean frontend-dev frontend-build run

# Default database path (can be overridden)
DB_PATH ?= /home/manuel/workspaces/2025-10-16/add-gpt5-responses-to-geppetto/geppetto/ttmp/2025-10-23/git-history-and-code-index.db
PORT ?= 8080

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

frontend-install: ## Install frontend dependencies
	cd frontend && npm install

frontend-dev: ## Start frontend dev server (Vite)
	cd frontend && npm run dev

frontend-build: ## Build frontend for production
	cd frontend && npm run build

dev: ## Run backend in development mode (needs frontend-dev in another terminal)
	go run main.go --db $(DB_PATH) --dev --port $(PORT)

build: frontend-build ## Build complete application
	go build -o pr-history-code-browser main.go

run: ## Run production binary
	./pr-history-code-browser --db $(DB_PATH) --port $(PORT)

clean: ## Clean build artifacts
	rm -f pr-history-code-browser
	rm -rf frontend/dist
	rm -rf frontend/node_modules

test: ## Run Go tests
	go test ./...

# Development workflow helpers
setup: frontend-install ## Setup development environment
	@echo "Setup complete! Run 'make dev' in one terminal and 'make frontend-dev' in another."

start: build run ## Build and run production server

