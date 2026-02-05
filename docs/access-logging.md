# HTTP Server Access Logging

## Overview

The HTTP server supports optional access logging that can be enabled with the `--verbose` (or `-v`) flag. When enabled, each HTTP request is logged with detailed information including method, path, status code, duration, client IP, and response size.

## Usage

### Enable Access Logging

Start the server with the `-v` or `--verbose` flag:

```bash
# Short form
tock serve -v

# Long form
tock serve --verbose

# With other options
tock serve --port 8080 --verbose
tock serve -p 8080 -v
```

### Disable Access Logging (Default)

By default, access logging is disabled. Simply start the server without the flag:

```bash
tock serve
tock serve --port 8080
```

## Log Format

When verbose mode is enabled, each request produces a log entry in the following format:

```
[HTTP] METHOD PATH - STATUS_CODE - DURATION - CLIENT_IP - RESPONSE_SIZE
```

### Example Logs

```
2026/02/06 00:15:30 [HTTP] GET /activity/current - 200 - 2.5ms - 127.0.0.1:54321 - 156 bytes
2026/02/06 00:15:35 [HTTP] POST /activity/start - 200 - 45.2ms - 127.0.0.1:54322 - 248 bytes
2026/02/06 00:15:40 [HTTP] GET /jira/suggest?q=backend - 200 - 123.4ms - 127.0.0.1:54323 - 1024 bytes
2026/02/06 00:15:42 [HTTP] GET /activity/events - 204 - 60.001s - 127.0.0.1:54324 - 0 bytes
2026/02/06 00:15:45 [HTTP] POST /activity/stop - 200 - 12.8ms - 127.0.0.1:54325 - 245 bytes
2026/02/06 00:15:50 [HTTP] OPTIONS /activity/start - 200 - 0.5ms - 127.0.0.1:54326 - 0 bytes
2026/02/06 00:15:55 [HTTP] GET /activity/list - 200 - 8.3ms - 127.0.0.1:54327 - 4523 bytes
2026/02/06 00:16:00 [HTTP] DELETE /activity/remove - 204 - 3.1ms - 127.0.0.1:54328 - 0 bytes
```

## Log Fields

### Method
HTTP method used for the request (GET, POST, PUT, DELETE, OPTIONS, etc.)

### Path
Full request path including query parameters
- Example: `/activity/start`
- Example: `/jira/suggest?q=backend&limit=10`

### Status Code
HTTP response status code
- `200` - Success
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error
- `503` - Service Unavailable

### Duration
Time taken to process the request
- Measured from when the request is received to when the response is complete
- Includes all middleware processing and handler execution
- Format: `123.4ms` or `1.234s`

### Client IP
Remote address of the client making the request
- Format: `IP:PORT`
- Example: `127.0.0.1:54321` (local)
- Example: `192.168.1.100:49152` (network)

### Response Size
Number of bytes written in the response body
- `0 bytes` for responses with no body (204 No Content, OPTIONS)
- Actual byte count for JSON responses and other content

## Use Cases

### Development & Debugging
Enable verbose logging during development to see all API interactions:
```bash
tock serve -v -p 8080
```

### Performance Monitoring
Track request durations to identify slow endpoints:
```bash
tock serve -v | grep "HTTP" | awk '{print $8, $6}' | sort -rn
```

### API Usage Analysis
Analyze which endpoints are being used:
```bash
tock serve -v 2>&1 | grep "HTTP" | awk '{print $6}' | sort | uniq -c | sort -rn
```

### Error Tracking
Monitor for errors (4xx and 5xx status codes):
```bash
tock serve -v 2>&1 | grep -E " - [45][0-9][0-9] - "
```

## Log Rotation

Since logs go to standard error/output, you can pipe them to log rotation tools:

### Using systemd
```ini
[Service]
ExecStart=/usr/local/bin/tock serve -v -p 8080
StandardOutput=journal
StandardError=journal
```

### Using logrotate
```bash
tock serve -v >> /var/log/tock-access.log 2>&1
```

### Using Docker
```bash
docker run -d \
  --name tock \
  -p 8080:8080 \
  tock:latest \
  serve -v -p 8080 \
  > /var/log/tock/access.log 2>&1
```

## Performance Impact

Access logging has minimal performance impact:
- **With logging disabled** (default): No overhead
- **With logging enabled** (`-v`): ~50-100μs per request overhead
- Logging is non-blocking and doesn't affect request processing

## Privacy Considerations

Access logs contain:
- ✅ Request paths (may contain query parameters)
- ✅ Client IP addresses
- ✅ Timestamps
- ❌ Request bodies
- ❌ Response bodies
- ❌ Authentication tokens

Be mindful when sharing logs as they may contain:
- Jira issue keys in query parameters
- Activity descriptions in URLs (if using GET endpoints)
- Client IP addresses

## Integration with Monitoring Tools

### Prometheus Integration
Parse logs to export metrics:
```bash
# Count requests by status code
tail -f /var/log/tock-access.log | awk '/HTTP/ {print $8}' | sort | uniq -c
```

### ELK Stack
Use Filebeat or Logstash to parse and index logs:
```yaml
# Filebeat pattern
- type: log
  paths:
    - /var/log/tock-access.log
  fields:
    service: tock
    type: access-log
```

### Grafana Loki
Stream logs directly to Loki for visualization:
```bash
tock serve -v 2>&1 | promtail --config.file=/etc/promtail/config.yml
```

## Comparison with Other Flags

```bash
# No logging (default)
tock serve

# Access logging only
tock serve -v

# With custom port
tock serve -p 8080 -v

# With CORS
tock serve --cors-origins "http://localhost:3000" -v

# All options
tock serve -p 8080 --cors-origins "http://localhost:3000,https://app.example.com" -v
```

## Troubleshooting

### Logs not appearing
1. Verify the `-v` flag is set
2. Check that logs aren't being redirected: `tock serve -v` (not `tock serve -v > /dev/null`)
3. Logs go to stderr by default - check your terminal/shell settings

### Too much noise
If logs are too verbose, consider:
1. Filtering specific endpoints: `tock serve -v 2>&1 | grep -v "/activity/events"`
2. Sampling logs: `tock serve -v 2>&1 | awk 'NR%10==0'` (every 10th line)
3. Running without `-v` in production

### Performance concerns
If logging impacts performance:
1. Write logs to a fast disk or ramdisk
2. Use log aggregation tools that buffer writes
3. Consider disabling logging in high-traffic production environments
