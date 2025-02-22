---
Title: Google Maps Directions API Commands
Slug: maps-directions
Short: Detailed guide for the Directions API commands
Topics:
  - maps
  - directions
  - routing
Commands:
  - directions
  - optimize-route
Flags:
  - origin
  - destination
  - mode
  - waypoints
  - avoid
  - units
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Directions API Commands

The Directions API commands help you find routes between locations and optimize multi-stop journeys.

## Basic Directions

Get directions between two points:

```bash
maps directions \
  --origin "Times Square, NY" \
  --destination "Central Park, NY" \
  --mode walking \
  --avoid "highways,tolls" \
  --units metric
```

### Parameters

- `origin` (required): Starting location (address or lat,lng)
- `destination` (required): Ending location (address or lat,lng)
- `mode`: Travel mode (driving, walking, bicycling, transit)
- `waypoints`: Intermediate stops
- `avoid`: Features to avoid (tolls, highways, ferries)
- `units`: Unit system (metric, imperial)

### Output Fields

For each route:
- Summary information
  - `route_number`: Route identifier
  - `summary`: Route description
  - `distance`: Total distance
  - `duration`: Total duration

For each step:
- `step_number`: Step sequence
- `instruction`: Navigation instruction
- `distance`: Step distance
- `duration`: Step duration
- `start_location`: Starting coordinates
- `end_location`: Ending coordinates
- `travel_mode`: Mode of travel

## Route Optimization

Find the best route through multiple stops:

```bash
maps optimize-route \
  --start "Hotel Address" \
  --attractions "Museum 1,Park 2,Restaurant 3" \
  --mode walking \
  --time-per-stop 90
```

### Parameters

- `start` (required): Starting location
- `attractions` (required): List of places to visit
- `mode`: Travel mode
- `time-per-stop`: Minutes to spend at each location
- `return-to-start`: Whether to return to starting point

### Algorithm

The route optimization:
1. Gets coordinates for all locations
2. Builds a distance/time matrix
3. Solves the Traveling Salesman Problem
4. Considers:
   - Opening hours
   - Visit durations
   - Travel times
   - User preferences

### Output

- Optimized itinerary
  - Visit sequence
  - Arrival times
  - Travel segments
  - Total duration and distance

## Implementation Details

The commands use the Google Maps Directions API:

```go
// Basic directions
req := &maps.DirectionsRequest{
    Origin:      origin,
    Destination: destination,
    Mode:        maps.Mode(mode),
    Units:       maps.Units(units),
}

// Route optimization
type OptimizeRouteSettings struct {
    Start       string   `glazed.parameter:"start"`
    Attractions []string `glazed.parameter:"attractions"`
    Mode        string   `glazed.parameter:"mode"`
    TimePerStop int      `glazed.parameter:"time-per-stop"`
}
```

## Common Features

1. Output Formatting
   ```bash
   # Get JSON output
   maps directions --origin "A" --destination "B" --output-format json
   
   # Select specific fields
   maps directions --origin "A" --destination "B" --fields "distance,duration"
   ```

2. Error Handling
   - Invalid location handling
   - No route found cases
   - API error management

3. Debugging
   ```bash
   # Enable debug logging
   maps --log-level debug directions --origin "A" --destination "B"
   ```

## Best Practices

1. Location Formats
   - Use precise coordinates when possible
   - Provide well-formatted addresses
   - Consider geocoding addresses first

2. Route Optimization
   - Limit the number of stops (API restrictions)
   - Consider time windows for attractions
   - Account for traffic conditions

3. Performance
   - Cache frequent routes
   - Use batch requests when possible
   - Implement proper error handling

For more information:
- [Overview](01-overview.md)
- [Places API](02-places.md)
- [Advanced Usage](04-advanced.md) 