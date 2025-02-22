---
Title: Google Maps CLI Overview
Slug: maps-overview
Short: Overview of the Google Maps CLI tools and their capabilities
Topics:
  - maps
  - overview
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

# Google Maps CLI Overview

The Google Maps CLI tools provide a powerful command-line interface to interact with various Google Maps APIs. This overview will help you understand the available features and how to get started.

## Available Commands

The CLI provides two main command groups:

### Places API Commands
- `maps places search`: Search for places using text queries
- `maps places details`: Get detailed information about a place
- `maps places nearby`: Find places near a location

### Directions API Commands
- `maps directions`: Get directions between locations
- `maps optimize-route`: Find optimal routes between multiple locations

## Getting Started

1. Set up your API key:
   ```bash
   export GOOGLE_MAPS_API_KEY="your-api-key"
   ```

2. Try a basic search:
   ```bash
   maps places search --query "coffee shops" --location "40.7128,-74.0060"
   ```

3. Get directions:
   ```bash
   maps directions --origin "Times Square" --destination "Central Park"
   ```

## Next Steps

- Read the detailed command documentation
- Check out the tutorials for common use cases
- Explore advanced features and integrations

For more information:
- [Places API Commands](02-places.md)
- [Directions API Commands](03-directions.md)
- [Advanced Usage](04-advanced.md) 