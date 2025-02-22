---
Title: Google Maps Places API Commands
Slug: maps-places
Short: Detailed guide for the Places API commands
Topics:
  - maps
  - places
Commands:
  - places search
  - places details
  - places nearby
Flags:
  - query
  - location
  - radius
  - type
  - place-id
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Places API Commands

The Places API commands allow you to search for and get information about places using the Google Maps Places API.

## Search Command

Search for places using text queries and filters:

```bash
maps places search \
  --query "museums in Manhattan" \
  --location "40.7128,-74.0060" \
  --radius 5000 \
  --type museum
```

### Parameters

- `query` (required): Text to search for
- `location`: Location in lat,lng format
- `radius`: Search radius in meters (default: 1500)
- `type`: Type of place (e.g., restaurant, museum)

### Output Fields

- `name`: Place name
- `address`: Formatted address
- `place_id`: Unique place identifier
- `rating`: Average rating (0-5)
- `user_ratings_total`: Number of ratings
- `types`: Place types

## Details Command

Get detailed information about a specific place:

```bash
maps places details --place-id "ChIJN1t_tDeuEmsRUsoyG83frY4"
```

### Parameters

- `place-id` (required): Google Maps Place ID

### Output Fields

- Basic information (name, address)
- Contact details (phone, website)
- Ratings and reviews
- Opening hours
- Additional attributes

## Nearby Command

Find places near a specific location:

```bash
maps places nearby \
  --location "40.7128,-74.0060" \
  --radius 1000 \
  --type restaurant \
  --keyword "pizza"
```

### Parameters

- `location` (required): Location in lat,lng format
- `radius`: Search radius in meters (default: 1500)
- `type`: Type of place
- `keyword`: Additional search term

### Output Fields

- `name`: Place name
- `address`: Vicinity address
- `place_id`: Unique identifier
- `rating`: Average rating
- `user_ratings_total`: Number of ratings
- `types`: Place types

## Common Features

All commands support:

1. Output Formatting
   ```bash
   # JSON output
   maps places search --query "cafes" --output-format json
   
   # Custom fields
   maps places details --place-id "ID" --fields name,rating
   ```

2. Error Handling
   - Clear error messages for invalid parameters
   - Proper HTTP error handling
   - Rate limit handling

3. Debugging
   ```bash
   # Enable debug logging
   maps --log-level debug places search --query "cafes"
   ```

## Implementation Details

The commands use the Google Maps Go Client Library:

```go
// Search request
req := &maps.TextSearchRequest{
    Query:  query,
    Radius: radius,
}

// Details request
req := &maps.PlaceDetailsRequest{
    PlaceID: placeID,
}

// Nearby request
req := &maps.NearbySearchRequest{
    Location: &maps.LatLng{Lat: lat, Lng: lng},
    Radius:   radius,
}
```

For more information:
- [Overview](01-overview.md)
- [Directions API](03-directions.md)
- [Advanced Usage](04-advanced.md) 