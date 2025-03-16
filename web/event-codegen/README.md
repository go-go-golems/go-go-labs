# Mouse Event Code Generator

A simple web application that generates Ruby code for mouse event handling. Built with Go and HTMX.

## Features

- Generate Ruby event classes with customizable fields and features
- Live code preview with syntax highlighting
- Download generated code
- Vim mode support in the code editor
- HTMX-powered instant updates

## Running the Application

1. Make sure you have Go installed on your system
2. Clone this repository
3. Navigate to the `web/event-codegen` directory
4. Run the application:

```bash
go run main.go
```

5. Open your browser and visit `http://localhost:8080`

## Usage

1. Configure your event by selecting the desired fields and features
2. The code will be generated automatically as you make changes
3. Use the "Download Ruby Code" button to download the generated code
4. Toggle Vim mode if desired for the code editor

## Structure

- `main.go` - Go backend server with code generation logic
- `templates/index.html` - HTMX-powered frontend 