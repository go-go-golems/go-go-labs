---
Title: Getting Started with Google Maps CLI
Slug: maps-getting-started
Short: A step-by-step guide to get started with the Google Maps CLI tools
Topics:
  - maps
  - tutorial
Commands:
  - places
  - directions
Flags:
  - api-key
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

# Getting Started with Google Maps CLI

This tutorial will guide you through setting up and using the Google Maps CLI tools for common tasks.

## Prerequisites

1. Install the CLI:
   ```bash
   go install github.com/go-go-golems/go-go-labs/cmd/maps@latest
   ```

2. Get a Google Maps API key:
   - Visit the [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project or select an existing one
   - Enable the required APIs:
     - Places API
     - Directions API
     - Distance Matrix API

3. Set up your API key:
   ```bash
   export GOOGLE_MAPS_API_KEY="your-api-key"
   ```

## First Steps

### 1. Verify Installation

Check that everything is working:

```bash
# Check version and available commands
maps --version
maps --help

# Test API key
maps places search --query "test" --location "40.7128,-74.0060"
```

### 2. Basic Place Search

Find coffee shops in New York:

```bash
# Search for coffee shops
maps places search \
  --query "coffee shops" \
  --location "40.7128,-74.0060" \
  --radius 1000

# Get JSON output
maps places search \
  --query "coffee shops" \
  --location "40.7128,-74.0060" \
  --output-format json
```

### 3. Get Place Details

Get more information about a specific place:

```bash
# First, search for a place
maps places search --query "Statue of Liberty"

# Then, get details using the place ID
maps places details --place-id "PLACE_ID_FROM_SEARCH"
```

### 4. Find Nearby Places

Discover places near a location:

```bash
# Find restaurants near Times Square
maps places nearby \
  --location "40.7580,-73.9855" \
  --type restaurant \
  --radius 500

# Find multiple types of places
maps places nearby \
  --location "40.7580,-73.9855" \
  --type "restaurant|cafe|bar" \
  --radius 500
```

### 5. Get Directions

Find routes between locations:

```bash
# Get walking directions
maps directions \
  --origin "Times Square, NY" \
  --destination "Central Park, NY" \
  --mode walking

# Add waypoints
maps directions \
  --origin "Times Square, NY" \
  --destination "Central Park, NY" \
  --waypoints "Rockefeller Center,NY" \
  --mode walking
```

## Common Tasks

### Creating a Restaurant List

Create a list of highly-rated restaurants:

```bash
# Search for restaurants
maps places search \
  --query "restaurants" \
  --location "40.7128,-74.0060" \
  --output-format json | \
  jq 'map(select(.rating >= 4.5))'

# Save to file
maps places search \
  --query "restaurants" \
  --location "40.7128,-74.0060" \
  --output-format json \
  --output-file restaurants.json
```

### Planning a Walking Tour

Create a walking tour of attractions:

```bash
# Find attractions
maps places nearby \
  --location "40.7128,-74.0060" \
  --type tourist_attraction \
  --radius 2000

# Get optimized route
maps optimize-route \
  --start "Times Square, NY" \
  --attractions "Museum1,Museum2,Park1" \
  --mode walking \
  --time-per-stop 90
```

## Next Steps

1. Explore more commands:
   - Check the [Places API documentation](../topics/02-places.md)
   - Learn about [Directions API](../topics/03-directions.md)
   - Try [advanced features](../topics/04-advanced.md)

2. Customize output:
   - Try different output formats (JSON, YAML, table)
   - Create custom templates
   - Filter and sort results

3. Integrate with scripts:
   - Create automation scripts
   - Process and analyze data
   - Build custom workflows

For more information:
- [Overview](../topics/01-overview.md)
- [Places API Commands](../topics/02-places.md)
- [Directions API Commands](../topics/03-directions.md)
- [Advanced Usage](../topics/04-advanced.md) 