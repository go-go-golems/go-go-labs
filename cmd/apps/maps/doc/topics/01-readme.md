---
Title: Google Maps CLI Tools
Slug: readme
Short: A comprehensive guide to using and extending the Google Maps CLI tools
Topics:
  - maps
  - places
  - directions
  - routing
Commands:
  - places
  - directions
Flags:
  - api-key
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Google Maps CLI Tools

This guide provides a comprehensive overview of the Google Maps CLI tools, including usage examples, implementation details, and best practices for extending the functionality.

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Places API Commands](#places-api-commands)
4. [Directions API Commands](#directions-api-commands)
5. [Implementing New Commands](#implementing-new-commands)
6. [Best Practices](#best-practices)
7. [Advanced Usage](#advanced-usage)

## Overview

The Google Maps CLI tools provide a command-line interface to various Google Maps APIs, allowing you to:
- Search for places and points of interest
- Get detailed information about specific locations
- Find nearby places based on location and criteria
- Get directions between locations
- Optimize routes for multiple destinations

### Architecture

The tools are built using:
- Google Maps Go Client Library (`googlemaps.github.io/maps`)
- Cobra for CLI framework
- Glazed for parameter handling and output formatting
- Zerolog for structured logging

## Getting Started

### Prerequisites

1. Google Maps API Key
   ```bash
   # Set your API key as an environment variable
   export GOOGLE_MAPS_API_KEY="your-api-key"
   
   # Or provide it via command line flag
   maps --api-key="your-api-key" [command]
   ```

2. Required API Services:
   - Places API
   - Directions API
   - Distance Matrix API (for route optimization)

### Basic Usage

```bash
# Search for places
maps places search --query "coffee shops" --location "40.7128,-74.0060" --radius 1000

# Get place details
maps places details --place-id "ChIJN1t_tDeuEmsRUsoyG83frY4"

# Find nearby places
maps places nearby --location "40.7128,-74.0060" --type restaurant --radius 500

# Get directions
maps directions \
  --origin "Times Square, NY" \
  --destination "Central Park, NY" \
  --mode walking
```

## Places API Commands

The Places API integration provides three main commands: `search`, `details`, and `nearby`.

### Place Search

The `search` command allows you to find places using text queries and filters:

```bash
maps places search \
  --query "museums in Manhattan" \
  --location "40.7128,-74.0060" \
  --radius 5000 \
  --type museum
```

Implementation Details:
```go
type SearchSettings struct {
    Query    string `glazed.parameter:"query"`
    Location string `glazed.parameter:"location"`
    Radius   int    `glazed.parameter:"radius"`
    Type     string `glazed.parameter:"type"`
}

// Uses maps.TextSearchRequest under the hood
req := &maps.TextSearchRequest{
    Query:  settings.Query,
    Radius: uint(settings.Radius),
    Type:   maps.PlaceType(settings.Type),
}
```

### Place Details

Get comprehensive information about a specific place:

```bash
maps places details --place-id "ChIJN1t_tDeuEmsRUsoyG83frY4"
```

The command returns:
- Basic information (name, address, coordinates)
- Contact details (phone, website)
- Ratings and reviews
- Opening hours
- Photos (if available)
- Additional attributes (price level, etc.)

Implementation:
```go
type DetailsSettings struct {
    PlaceID string `glazed.parameter:"place-id"`
}

// Uses maps.PlaceDetailsRequest
req := &maps.PlaceDetailsRequest{
    PlaceID: settings.PlaceID,
}
```

### Nearby Search

Find places near a specific location:

```bash
maps places nearby \
  --location "40.7128,-74.0060" \
  --radius 1000 \
  --type restaurant \
  --keyword "pizza"
```

Implementation:
```go
type NearbySettings struct {
    Location string `glazed.parameter:"location"`
    Radius   int    `glazed.parameter:"radius"`
    Type     string `glazed.parameter:"type"`
    Keyword  string `glazed.parameter:"keyword"`
}

// Uses maps.NearbySearchRequest
req := &maps.NearbySearchRequest{
    Location: &maps.LatLng{Lat: lat, Lng: lng},
    Radius:   uint(settings.Radius),
    Type:     maps.PlaceType(settings.Type),
    Keyword:  settings.Keyword,
}
```

## Directions API Commands

The Directions API integration provides routing capabilities between locations.

### Basic Directions

Get directions between two points:

```bash
maps directions \
  --origin "Times Square, NY" \
  --destination "Central Park, NY" \
  --mode walking \
  --avoid "highways,tolls"
```

Implementation:
```go
type DirectionsSettings struct {
    Origin      string   `glazed.parameter:"origin"`
    Destination string   `glazed.parameter:"destination"`
    Mode        string   `glazed.parameter:"mode"`
    Waypoints   []string `glazed.parameter:"waypoints"`
    Avoid       []string `glazed.parameter:"avoid"`
    Units       string   `glazed.parameter:"units"`
}

// Uses maps.DirectionsRequest
req := &maps.DirectionsRequest{
    Origin:      settings.Origin,
    Destination: settings.Destination,
    Mode:        maps.Mode(strings.ToUpper(settings.Mode)),
    Units:       maps.Units(strings.ToUpper(settings.Units)),
}
```

### Route Optimization

For visiting multiple locations efficiently:

```bash
maps optimize-route \
  --start "Hotel Address" \
  --attractions "Attraction1,Attraction2,Attraction3" \
  --mode walking \
  --time-per-stop 90
```

This command:
1. Retrieves place details for each attraction
2. Builds a distance/time matrix
3. Solves the Traveling Salesman Problem
4. Returns an optimized itinerary

## Implementing New Commands

To add a new command, follow these steps:

1. **Create Settings Struct**
   ```go
   type NewCommandSettings struct {
       Param1 string `glazed.parameter:"param1"`
       Param2 int    `glazed.parameter:"param2"`
   }
   ```

2. **Create Command Struct**
   ```go
   type NewCommand struct {
       *cmds.CommandDescription
       settings NewCommandSettings
   }
   ```

3. **Implement Command Factory**
   ```go
   func NewCustomCommand() (*cobra.Command, error) {
       glazedLayer, err := settings.NewGlazedParameterLayers()
       if err != nil {
           return nil, err
       }
       
       cmd := &NewCommand{
           CommandDescription: cmds.NewCommandDescription(
               "command-name",
               cmds.WithShort("Short description"),
               cmds.WithLong("Detailed description"),
               cmds.WithFlags(...),
               cmds.WithLayersList(glazedLayer),
           ),
       }
       
       return cli.BuildCobraCommandFromGlazeCommand(cmd)
   }
   ```

4. **Implement RunIntoGlazeProcessor**
   ```go
   func (c *NewCommand) RunIntoGlazeProcessor(
       ctx context.Context,
       parsedLayers *layers.ParsedLayers,
       gp middlewares.Processor,
   ) error {
       if err := parsedLayers.InitializeStruct(layers.DefaultSlug, &c.settings); err != nil {
           return err
       }
       
       // Implementation here
       
       return nil
   }
   ```

## Best Practices

1. **Error Handling**
   - Always validate input parameters
   - Provide clear error messages
   - Use structured logging for debugging

2. **Output Formatting**
   - Use Glazed's structured output
   - Support multiple output formats (text, JSON, etc.)
   - Include relevant metadata

3. **Performance**
   - Cache frequently used data
   - Batch API requests when possible
   - Use appropriate timeouts

4. **Testing**
   - Write unit tests for parameter validation
   - Mock API responses for testing
   - Include integration tests

## Advanced Usage

### Custom Output Formatting

```bash
# Output as JSON
maps places search --query "museums" --output-format json

# Select specific fields
maps places details --place-id "ID" --fields "name,rating,address"

# Sort and filter results
maps places nearby --location "..." --sort-by rating --min-rating 4.0
```

### Batch Processing

```bash
# Process multiple places
maps places batch-details --place-ids "id1,id2,id3"

# Export results to file
maps places search --query "hotels" --output-file hotels.json
```

### Integration with Other Tools

```bash
# Pipe results to jq
maps places search --query "restaurants" --output-format json | jq .rating

# Use in scripts
places_json=$(maps places nearby --location "..." --output-format json)
```

For more information:
- [Google Maps API Documentation](https://developers.google.com/maps/documentation)
- [Go Client Library Reference](https://pkg.go.dev/googlemaps.github.io/maps)
- [Glazed Documentation](https://github.com/go-go-golems/glazed) 