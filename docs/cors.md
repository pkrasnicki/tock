# CORS Configuration

The Tock HTTP API includes built-in CORS (Cross-Origin Resource Sharing) support to allow web applications from different origins to interact with the API.

## Default Configuration

By default, the server uses a permissive CORS configuration:

- **Allowed Origins**: `*` (all origins)
- **Allowed Methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- **Allowed Headers**: `Accept`, `Authorization`, `Content-Type`, `X-CSRF-Token`
- **Exposed Headers**: `Link`
- **Allow Credentials**: `false`
- **Max Age**: `300` seconds

## Customizing CORS Origins

You can restrict which origins are allowed to access the API using the `--cors-origins` flag:

```bash
# Allow a single origin
tock serve --cors-origins "https://example.com"

# Allow multiple origins (comma-separated)
tock serve --cors-origins "https://example.com,https://app.example.com,http://localhost:3000"

# Allow all origins (default behavior)
tock serve
```

## Programmatic Configuration

If you're integrating the HTTP handlers into your own application, you can customize the CORS configuration:

```go
import (
    httpAdapter "github.com/kriuchkov/tock/internal/adapters/http"
)

// Create custom CORS config
corsConfig := httpAdapter.CORSConfig{
    AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    ExposedHeaders:   []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           600,
}

// Create handler with custom CORS
handler := httpAdapter.NewHandlerWithCORS(service, corsConfig)
handler.RegisterRoutes()
```

## CORS Headers

The server automatically adds the following headers to responses:

- `Access-Control-Allow-Origin`: The allowed origin(s)
- `Access-Control-Allow-Methods`: Allowed HTTP methods
- `Access-Control-Allow-Headers`: Allowed request headers
- `Access-Control-Expose-Headers`: Headers exposed to the client
- `Access-Control-Allow-Credentials`: Whether credentials are allowed
- `Access-Control-Max-Age`: How long preflight results can be cached

## Preflight Requests

The server automatically handles preflight OPTIONS requests. When a browser sends a preflight request, the server responds with appropriate CORS headers and a `204 No Content` status.

## Security Considerations

- **Production deployments**: Always specify explicit origins rather than using `*`
- **Credentials**: Only enable `AllowCredentials` when necessary and with specific origins (not `*`)
- **Headers**: Only allow headers that your API actually uses
- **Methods**: Restrict methods to only those your API supports

## Examples

### Development Setup

For local development with a frontend on `localhost:3000`:

```bash
tock serve --port 8080 --cors-origins "http://localhost:3000"
```

### Production Setup

For production with specific domains:

```bash
tock serve --port 8080 --cors-origins "https://app.example.com,https://www.example.com"
```

### Testing with curl

Test CORS headers with a preflight request:

```bash
curl -X OPTIONS http://localhost:8080/activity/list \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -v
```

Test actual request with CORS:

```bash
curl http://localhost:8080/activity/current \
  -H "Origin: http://localhost:3000" \
  -v
```
