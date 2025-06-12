---
slug: queries
title: Writing YAML Queries
---

# Writing YAML Queries

The Datadog CLI uses YAML files to define reusable query templates. This allows for parameterized queries that can be shared and version-controlled.

## Basic Query Structure

```yaml
name: my_query
short: Brief description of the query
flags:
  - name: service
    type: string
    help: Service name to filter by
  - name: from
    type: string
    default: "-1h"
    help: Start time
query: |
  service:{{ .service | ddLike }} AND status:error
  @timestamp:[{{ .from | ddDateTime }} TO {{ .to | ddDateTime }}]
subqueries:
  sort: "desc"
```

## Template Helpers

- `ddLike` - Safely quote strings for search
- `ddDateTime` - Format times for Datadog API
- `ddStringIn` - Format string lists for IN queries
- `ddFacet` - Format field names for faceting

## Examples

### Service Error Logs
```yaml
name: service_errors
short: Get error logs for a specific service
flags:
  - name: service
    type: string
    required: true
query: service:{{ .service | ddLike }} AND status:error
```

### Time Range Query
```yaml
name: recent_activity
short: Recent activity in time range
flags:
  - name: from
    type: string
    default: "-1h"
  - name: to
    type: string
    default: "now"
query: "@timestamp:[{{ .from | ddDateTime }} TO {{ .to | ddDateTime }}]"
```

## Running Queries

- **Built-in queries:** `datadog-cli logs <query-name>`
- **Custom files:** `datadog-cli logs run my-query.yaml`
- **Raw queries:** `datadog-cli logs query "service:web-api"`
