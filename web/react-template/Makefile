.PHONY: build frontend-build backend-build clean run dev

# Default target
all: build

# Build frontend assets
frontend-build:
	bun run build
	mkdir -p cmd/server/static
	cp -r dist/* cmd/server/static/

# Build the Go backend
backend-build: frontend-build
	go build -o server ./cmd/server

# Full build process
build: backend-build

# Clean built assets
clean:
	rm -rf dist
	rm -rf cmd/server/static
	rm -f server

# Run the server
run: build
	./server

# Run Vite development server (proxying API calls to Go backend)
dev:
	bun run dev 