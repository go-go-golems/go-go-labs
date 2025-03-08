---
Title: Planning a Sightseeing Tour
Slug: maps-sightseeing
Short: Example of using the Google Maps CLI to plan a sightseeing tour
Topics:
  - maps
  - examples
  - tourism
Commands:
  - places
  - directions
  - optimize-route
Flags:
  - location
  - type
  - radius
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Example
---

# Planning a Sightseeing Tour with Google Maps CLI

This example shows how to use the Google Maps CLI to plan an efficient sightseeing tour of New York City.

## Step 1: Find Tourist Attractions

First, let's find popular tourist attractions in the area:

```bash
# Search for tourist attractions near Times Square
maps places nearby \
  --location "40.7580,-73.9855" \
  --type tourist_attraction \
  --radius 3000 \
  --output-format json > attractions.json

# Filter for highly-rated attractions
cat attractions.json | jq 'map(select(.rating >= 4.0))'
```

## Step 2: Get Attraction Details

Get detailed information about each attraction:

```bash
# Create a script to get details for each attraction
#!/bin/bash

# Read place IDs from attractions.json
place_ids=$(cat attractions.json | jq -r '.[].place_id')

# Get details for each place
for id in $place_ids; do
  maps places details \
    --place-id "$id" \
    --fields "name,rating,opening_hours,address" \
    --output-format json
done > attraction_details.json
```

## Step 3: Plan the Route

Create an optimized route through the attractions:

```bash
# Starting from Times Square
maps optimize-route \
  --start "Times Square, NY" \
  --attractions "Empire State Building,Central Park,Statue of Liberty" \
  --mode walking \
  --time-per-stop 90 \
  --output-format json > route.json

# Get detailed directions
maps directions \
  --origin "Times Square, NY" \
  --destination "Central Park, NY" \
  --waypoints "Empire State Building,NY" \
  --mode walking \
  --avoid highways \
  --units metric \
  --output-format json > directions.json
```

## Step 4: Create an Itinerary

Process the route and create a detailed itinerary:

```bash
#!/bin/bash

# Function to format time
format_time() {
  date -d "@$1" +"%H:%M"
}

# Start time (10:00 AM)
start_time=$(date -d "10:00" +%s)
current_time=$start_time

# Read the route
while IFS= read -r stop; do
  name=$(echo "$stop" | jq -r '.name')
  duration=$(echo "$stop" | jq -r '.duration')
  
  # Print arrival time and location
  echo "$(format_time $current_time): Arrive at $name"
  
  # Add visit duration (90 minutes)
  current_time=$((current_time + 5400))
  echo "$(format_time $current_time): Leave $name"
  
  # Add travel time to next stop
  current_time=$((current_time + duration))
done < <(jq -c '.stops[]' route.json)
```

## Step 5: Export Results

Create a shareable format:

```bash
# Create a markdown summary
cat << 'EOF' > itinerary.md
# New York City Sightseeing Tour

## Attractions
$(jq -r '.[].name' attractions.json | sed 's/^/- /')

## Route Details
$(jq -r '.summary' route.json)

## Step-by-Step Directions
$(jq -r '.steps[].instruction' directions.json | sed 's/^/1. /')
EOF

# Create a JSON file for further processing
jq -s '{ 
  attractions: .[0],
  route: .[1],
  directions: .[2]
}' attractions.json route.json directions.json > tour.json
```

## Complete Script

Here's a complete script that puts it all together:

```bash
#!/bin/bash
set -euo pipefail

# Configuration
START_LOCATION="Times Square, NY"
RADIUS=3000  # meters
MIN_RATING=4.0
VISIT_DURATION=90  # minutes

# Find attractions
echo "Finding attractions..."
maps places nearby \
  --location "40.7580,-73.9855" \
  --type tourist_attraction \
  --radius $RADIUS \
  --output-format json | \
  jq "map(select(.rating >= $MIN_RATING))" > attractions.json

# Get attraction details
echo "Getting attraction details..."
place_ids=$(jq -r '.[].place_id' attractions.json)
for id in $place_ids; do
  maps places details \
    --place-id "$id" \
    --fields "name,rating,opening_hours,address" \
    --output-format json
done > attraction_details.json

# Create attraction list
attractions=$(jq -r '.[].name' attractions.json | paste -sd,)

# Optimize route
echo "Planning route..."
maps optimize-route \
  --start "$START_LOCATION" \
  --attractions "$attractions" \
  --mode walking \
  --time-per-stop $VISIT_DURATION \
  --output-format json > route.json

# Get detailed directions
echo "Getting directions..."
waypoints=$(jq -r '.stops[:-1].name' route.json | paste -sd,)
final_stop=$(jq -r '.stops[-1].name' route.json)

maps directions \
  --origin "$START_LOCATION" \
  --destination "$final_stop" \
  --waypoints "$waypoints" \
  --mode walking \
  --avoid highways \
  --units metric \
  --output-format json > directions.json

# Create itinerary
echo "Creating itinerary..."
./create_itinerary.sh > itinerary.md

echo "Done! Check itinerary.md for your tour plan."
```

## Usage

1. Save the script as `plan_tour.sh`
2. Make it executable:
   ```bash
   chmod +x plan_tour.sh
   ```

3. Run the script:
   ```bash
   ./plan_tour.sh
   ```

4. Check the results:
   ```bash
   cat itinerary.md
   ```

## Output Example

The script generates a detailed itinerary like this:

```markdown
# New York City Sightseeing Tour

## Schedule
10:00 - Arrive at Times Square
11:30 - Empire State Building
13:00 - Lunch Break
14:00 - Central Park
15:30 - Metropolitan Museum
17:00 - Tour Ends

## Route Summary
- Total Distance: 5.2 km
- Walking Time: 65 minutes
- Attractions: 4
- Duration: 7 hours

## Step-by-Step Directions
1. Head northeast on Broadway
2. Turn right onto W 42nd St
3. ...
```

For more information:
- [Places API Commands](../topics/02-places.md)
- [Directions API Commands](../topics/03-directions.md)
- [Advanced Usage](../topics/04-advanced.md) 