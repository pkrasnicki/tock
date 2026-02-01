# HTTP API Examples

This directory contains example HTTP requests for the Tock API.

## Prerequisites

Start the Tock HTTP server:

```bash
tock serve --port 8080
```

## Using the Examples

### With VS Code REST Client Extension

Install the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension for VS Code, then open any `.http` file and click "Send Request" above each request.

### With curl

#### Start Activity

```bash
# POST with JSON
curl -X POST http://localhost:8080/activity/start \
  -H "Content-Type: application/json" \
  -d '{"Description": "Working on API", "Project": "TOCK"}'

# GET with query params
curl "http://localhost:8080/activity/start?project=TOCK&description=Working%20on%20API"
```

#### Stop Activity

```bash
# Stop current activity
curl http://localhost:8080/activity/stop

# Stop with specific time
curl -X POST http://localhost:8080/activity/stop \
  -H "Content-Type: application/json" \
  -d '{"EndTime": "2026-01-31T12:30:00Z"}'
```

#### Add Completed Activity

```bash
curl -X POST http://localhost:8080/activity/add \
  -H "Content-Type: application/json" \
  -d '{
    "Description": "Morning standup",
    "Project": "TEAM",
    "StartTime": "2026-01-31T09:00:00Z",
    "EndTime": "2026-01-31T09:30:00Z"
  }'
```

#### List Activities

```bash
# All activities
curl http://localhost:8080/activity/list

# With filters
curl "http://localhost:8080/activity/list?from=2026-01-01T00:00:00Z&to=2026-01-31T23:59:59Z&project=TOCK"
```

#### Current Activities

```bash
curl http://localhost:8080/activity/current
```

#### Recent Activities

```bash
# Default (10)
curl http://localhost:8080/activity/recent

# Custom limit
curl "http://localhost:8080/activity/recent?limit=20"
```

#### Activity Report

```bash
# All activities
curl http://localhost:8080/activity/report

# With filters
curl "http://localhost:8080/activity/report?from=2026-01-01T00:00:00Z&to=2026-01-31T23:59:59Z"
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/activity/start` | GET/POST | Start a new activity |
| `/activity/stop` | GET/POST | Stop current activity |
| `/activity/add` | POST | Add a completed activity |
| `/activity/list` | GET | List activities with filters |
| `/activity/current` | GET | Get currently running activities |
| `/activity/recent` | GET | Get recent unique activities |
| `/activity/report` | GET | Get activity report with duration breakdown |

## Response Format

All endpoints return JSON responses with appropriate HTTP status codes:

- `200 OK` - Success
- `400 Bad Request` - Invalid request data
- `405 Method Not Allowed` - Wrong HTTP method
- `500 Internal Server Error` - Server error
