# Activity Events - Long-Polling API

## Overview

The Activity Events API provides real-time notifications for activity changes using long-polling. This allows clients to be immediately notified when activities are started, stopped, added, or removed without constantly polling the server.

## Endpoint

```
GET /activity/events
```

## How Long-Polling Works

1. **Client Request**: Client makes a GET request to `/activity/events`
2. **Server Wait**: Server keeps the connection open for up to 60 seconds
3. **Event Occurs**: If an activity event happens during this time, server responds immediately with event data
4. **Timeout**: If no event occurs within 60 seconds, server responds with `204 No Content`
5. **Client Reconnect**: Client immediately makes another request to continue listening

## Event Types

The API broadcasts four types of events:

### 1. Activity Started
Triggered when a new activity is started.

```json
{
  "type": "activity_started",
  "activity": {
    "id": "20240205120000",
    "description": "Working on feature X",
    "project": "MyProject",
    "start": "2024-02-05T12:00:00Z",
    "tags": ["development"]
  },
  "timestamp": "2024-02-05T12:00:00Z"
}
```

### 2. Activity Stopped
Triggered when an activity is stopped.

```json
{
  "type": "activity_stopped",
  "activity": {
    "id": "20240205120000",
    "description": "Working on feature X",
    "project": "MyProject",
    "start": "2024-02-05T12:00:00Z",
    "end": "2024-02-05T13:00:00Z",
    "tags": ["development"]
  },
  "timestamp": "2024-02-05T13:00:00Z"
}
```

### 3. Activity Added
Triggered when a past activity is added to the timeline.

```json
{
  "type": "activity_added",
  "activity": {
    "id": "20240205110000",
    "description": "Meeting",
    "project": "Admin",
    "start": "2024-02-05T11:00:00Z",
    "end": "2024-02-05T11:30:00Z",
    "tags": ["meeting"]
  },
  "timestamp": "2024-02-05T13:00:05Z"
}
```

### 4. Activity Removed
Triggered when an activity is deleted.

```json
{
  "type": "activity_removed",
  "timestamp": "2024-02-05T13:00:10Z"
}
```

## Response Codes

- `200 OK` - Event occurred and returned in response body
- `204 No Content` - Timeout reached with no events (client should reconnect)
- `405 Method Not Allowed` - Wrong HTTP method used (only GET allowed)
- `408 Request Timeout` - Client disconnected
- `503 Service Unavailable` - Events feature not enabled

## Client Implementation Example

### JavaScript/TypeScript
```javascript
async function listenForEvents() {
  while (true) {
    try {
      const response = await fetch('http://localhost:8080/activity/events');
      
      if (response.status === 200) {
        const event = await response.json();
        console.log('Activity event:', event);
        
        // Handle the event
        switch (event.type) {
          case 'activity_started':
            console.log('Activity started:', event.activity);
            break;
          case 'activity_stopped':
            console.log('Activity stopped:', event.activity);
            break;
          case 'activity_added':
            console.log('Activity added:', event.activity);
            break;
          case 'activity_removed':
            console.log('Activity removed');
            break;
        }
      } else if (response.status === 204) {
        // Timeout, reconnect immediately
        console.log('No events, reconnecting...');
      } else {
        console.error('Unexpected response:', response.status);
        await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5s on error
      }
    } catch (error) {
      console.error('Error listening for events:', error);
      await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5s on error
    }
  }
}

// Start listening
listenForEvents();
```

### Python
```python
import requests
import time

def listen_for_events():
    url = 'http://localhost:8080/activity/events'
    
    while True:
        try:
            response = requests.get(url, timeout=65)
            
            if response.status_code == 200:
                event = response.json()
                print(f"Activity event: {event['type']}")
                
                # Handle the event
                if event['type'] == 'activity_started':
                    print(f"Activity started: {event['activity']}")
                elif event['type'] == 'activity_stopped':
                    print(f"Activity stopped: {event['activity']}")
                elif event['type'] == 'activity_added':
                    print(f"Activity added: {event['activity']}")
                elif event['type'] == 'activity_removed':
                    print("Activity removed")
                    
            elif response.status_code == 204:
                print("No events, reconnecting...")
            else:
                print(f"Unexpected response: {response.status_code}")
                time.sleep(5)
                
        except requests.exceptions.RequestException as e:
            print(f"Error: {e}")
            time.sleep(5)

# Start listening
listen_for_events()
```

### curl (for testing)
```bash
# Listen for a single event
curl -i http://localhost:8080/activity/events

# Continuous listening (reconnects automatically)
while true; do
  curl -s http://localhost:8080/activity/events | jq .
  echo "Reconnecting..."
  sleep 1
done
```

## Architecture

### EventBroadcaster
The `EventBroadcaster` manages subscribers and broadcasts events to all connected clients:

- **Thread-safe**: Uses mutex locks for concurrent access
- **Auto-cleanup**: Removes stale subscribers automatically every 30 seconds
- **Buffer**: Each subscriber has a 10-event buffer to prevent blocking
- **Timeout**: Default 60-second timeout for long-polling requests

### Integration
Events are automatically broadcast when:
- `POST /activity/start` - Broadcasts `activity_started`
- `POST /activity/stop` - Broadcasts `activity_stopped`
- `POST /activity/add` - Broadcasts `activity_added`
- `DELETE /activity/remove` - Broadcasts `activity_removed`

## Configuration

The event broadcaster is automatically initialized with default settings:
- **Timeout**: 60 seconds
- **Buffer size**: 10 events per subscriber
- **Cleanup interval**: 30 seconds

To customize, use the appropriate constructor:
```go
// Default with events enabled
handler := http.NewHandler(service)

// Custom timeout
broadcaster := http.NewEventBroadcaster(30 * time.Second)
handler := http.NewHandlerWithEvents(service, broadcaster)

// With CORS and events
handler := http.NewHandlerWithCORSAndEvents(service, corsConfig, broadcaster)
```

## Best Practices

1. **Always reconnect**: After receiving an event or timeout, immediately reconnect
2. **Handle timeouts gracefully**: 204 responses are normal, not errors
3. **Implement exponential backoff**: On connection errors, wait before retrying
4. **Use unique subscriber IDs**: Each connection gets a unique ID automatically
5. **Monitor memory**: Stale connections are cleaned up automatically, but monitor your client connections

## Performance Considerations

- **Scalability**: Each connected client holds a goroutine and channel
- **Memory**: ~10KB per connected client (buffered channel)
- **Cleanup**: Automatic cleanup prevents memory leaks from abandoned connections
- **No polling**: More efficient than regular polling, but use WebSockets for very high-frequency updates

## Limitations

1. **One event per connection**: Each long-polling request returns at most one event
2. **No event history**: Only events that occur while connected are received
3. **60-second timeout**: Requests timeout after 60 seconds if no events occur
4. **Broadcast only**: All subscribers receive all events (no filtering)

## Future Enhancements

Potential improvements for consideration:
- Event filtering by project or activity type
- Configurable timeout per client
- Event history/replay for reconnecting clients
- WebSocket support for true push notifications
- Event batching for high-frequency updates
