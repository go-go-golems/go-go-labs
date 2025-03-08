---
Title: Advanced Google Maps CLI Usage
Slug: maps-advanced
Short: Advanced features and best practices for the Google Maps CLI tools
Topics:
  - maps
  - advanced
  - integration
Commands:
  - places
  - directions
Flags:
  - output-format
  - fields
  - sort-by
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Advanced Google Maps CLI Usage

This guide covers advanced features, integration patterns, and best practices for the Google Maps CLI tools.

## Output Formatting

### Glazed Output Options

All commands support Glazed's output formatting:

```bash
# JSON output
maps places search --query "museums" \
  --output-format json \
  --output-file museums.json

# Table format with specific columns
maps places nearby --location "40.7128,-74.0060" \
  --output-format table \
  --fields "name,rating,address"

# YAML output with sorting
maps places search --query "restaurants" \
  --output-format yaml \
  --sort-by "rating:desc"
```

### Custom Templates

Create custom output templates:

```bash
# Create a template
cat > place.tmpl <<'EOF'
{{range .}}
Name: {{.name}}
Rating: {{.rating}}/5 ({{.user_ratings_total}} reviews)
Address: {{.address}}
{{end}}
EOF

# Use the template
maps places search --query "cafes" \
  --output-format template \
  --template-file place.tmpl
```

## Integration Patterns

### Shell Scripts

Integrate with shell scripts:

```bash
#!/bin/bash

# Find nearby attractions
attractions=$(maps places nearby \
  --location "$1" \
  --type "tourist_attraction" \
  --radius 5000 \
  --output-format json)

# Extract place IDs
place_ids=$(echo "$attractions" | jq -r '.[].place_id')

# Get details for each attraction
for id in $place_ids; do
  maps places details \
    --place-id "$id" \
    --fields "name,rating,opening_hours"
done
```

### Data Processing

Process and analyze results:

```bash
# Find high-rated restaurants
maps places search \
  --query "restaurants" \
  --location "40.7128,-74.0060" \
  --output-format json | \
  jq 'map(select(.rating >= 4.5))'

# Calculate average ratings
maps places nearby \
  --location "40.7128,-74.0060" \
  --type "cafe" \
  --output-format json | \
  jq '[.[].rating] | add/length'
```

## Performance Optimization

### Caching

Implement response caching:

```bash
# Cache directory
export MAPS_CACHE_DIR="$HOME/.cache/maps-cli"

# Cache duration
export MAPS_CACHE_TTL="24h"

# Use cached results when available
maps places search \
  --query "museums" \
  --use-cache \
  --cache-ttl "24h"
```

### Batch Processing

Process multiple items efficiently:

```bash
# Batch place details
maps places batch-details \
  --place-ids "id1,id2,id3" \
  --parallel 3

# Bulk geocoding
maps geocode batch \
  --addresses-file locations.txt \
  --output-format json
```

## Error Handling

### Retry Logic

Handle transient failures:

```bash
# Automatic retries
maps places search \
  --query "restaurants" \
  --retry-count 3 \
  --retry-delay 1s

# Custom error handling
maps places details \
  --place-id "ID" \
  --on-error "skip" \
  --error-output errors.log
```

### Rate Limiting

Respect API limits:

```bash
# Set request limits
maps places search \
  --query "hotels" \
  --qps-limit 10 \
  --quota-limit 1000

# Handle quota exceeded
maps places nearby \
  --location "..." \
  --on-quota-exceeded "wait"
```

## Extending Functionality

### Custom Commands

Create new commands:

```go
type CustomCommand struct {
    *cmds.CommandDescription
    settings CustomSettings
}

func NewCustomCommand() (*cobra.Command, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    cmd := &CustomCommand{
        CommandDescription: cmds.NewCommandDescription(
            "custom",
            cmds.WithShort("Custom command"),
            cmds.WithFlags(...),
            cmds.WithLayersList(glazedLayer),
        ),
    }
    
    return cli.BuildCobraCommandFromGlazeCommand(cmd)
}
```

### Middleware

Add custom middleware:

```go
func LoggingMiddleware(next middlewares.Processor) middlewares.Processor {
    return middlewares.ProcessorFunc(func(ctx context.Context, row *types.Row) error {
        log.Info().Interface("row", row).Msg("Processing row")
        return next.ProcessRow(ctx, row)
    })
}

// Use middleware
cmd.Use(LoggingMiddleware)
```

## Configuration

### Profiles

Create configuration profiles:

```yaml
# ~/.config/maps-cli/config.yaml
profiles:
  default:
    api_key: "${GOOGLE_MAPS_API_KEY}"
    output_format: table
    units: metric
  
  development:
    api_key: "${DEV_API_KEY}"
    log_level: debug
    retry_count: 3
  
  production:
    api_key: "${PROD_API_KEY}"
    cache_enabled: true
    qps_limit: 50
```

### Environment Variables

Use environment variables:

```bash
# API configuration
export GOOGLE_MAPS_API_KEY="your-key"
export MAPS_DEFAULT_LOCATION="40.7128,-74.0060"
export MAPS_OUTPUT_FORMAT="json"

# Debugging
export MAPS_LOG_LEVEL="debug"
export MAPS_DEBUG_HTTP="true"

# Performance
export MAPS_CACHE_ENABLED="true"
export MAPS_QPS_LIMIT="10"
```

## Testing

### Unit Tests

Write comprehensive tests:

```go
func TestPlacesSearch(t *testing.T) {
    tests := []struct {
        name     string
        query    string
        location string
        want     []Place
        wantErr  bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewSearchCommand()
            // Test implementation
        })
    }
}
```

### Integration Tests

Test with real API:

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Test with real API
    cmd := NewSearchCommand()
    // Test implementation
}
```

For more information:
- [Overview](01-overview.md)
- [Places API](02-places.md)
- [Directions API](03-directions.md) 