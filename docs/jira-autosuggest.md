# Jira Autosuggest Integration

## Overview

The Jira Autosuggest feature provides real-time issue suggestions as users type, making it easy to associate activities with Jira issues without manually looking up issue keys.

## Endpoint

```
GET /jira/suggest?q=<query>&limit=<limit>
```

## Configuration

To enable Jira autosuggest, configure your Jira credentials in `~/.config/tock/tock.yaml`:

```yaml
jira:
  url: "https://your-domain.atlassian.net"
  username: "your-email@example.com"
  api_token: "your-api-token"
```

### Getting a Jira API Token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a descriptive name (e.g., "Tock Integration")
4. Copy the token and add it to your config file

### Environment Variables

You can also configure Jira using environment variables:

```bash
export TOCK_JIRA_URL="https://your-domain.atlassian.net"
export TOCK_JIRA_USERNAME="your-email@example.com"
export TOCK_JIRA_API_TOKEN="your-api-token"
```

## Parameters

### Required
- **q** (query string): Search query, minimum 2 characters
  - Searches both issue keys and summaries
  - Case-insensitive
  - Supports partial matches

### Optional
- **limit** (integer): Maximum number of results to return
  - Default: 10
  - Minimum: 1
  - Maximum: 50

## Response Format

```json
{
  "query": "backend",
  "suggestions": [
    {
      "key": "BACKEND-123",
      "summary": "Implement user authentication"
    },
    {
      "key": "BACKEND-124",
      "summary": "Add API rate limiting"
    }
  ]
}
```

## Usage Examples

### Basic Search

Search for issues containing "backend":

```bash
curl "http://localhost:8080/jira/suggest?q=backend"
```

### Search by Issue Key

Find issues starting with a project key:

```bash
curl "http://localhost:8080/jira/suggest?q=PROJ"
```

### Limit Results

Get only 5 suggestions:

```bash
curl "http://localhost:8080/jira/suggest?q=api&limit=5"
```

### JavaScript/TypeScript Example

```javascript
async function getJiraSuggestions(query) {
  if (query.length < 2) {
    return [];
  }

  const response = await fetch(
    `http://localhost:8080/jira/suggest?q=${encodeURIComponent(query)}&limit=10`
  );

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data = await response.json();
  return data.suggestions;
}

// Usage in an autocomplete component
const input = document.getElementById('jira-issue');
let debounceTimer;

input.addEventListener('input', (e) => {
  clearTimeout(debounceTimer);
  const query = e.target.value;

  debounceTimer = setTimeout(async () => {
    if (query.length >= 2) {
      const suggestions = await getJiraSuggestions(query);
      displaySuggestions(suggestions);
    }
  }, 300); // Wait 300ms after user stops typing
});

function displaySuggestions(suggestions) {
  const dropdown = document.getElementById('suggestions');
  dropdown.innerHTML = suggestions
    .map(s => `
      <div class="suggestion" data-key="${s.key}">
        <strong>${s.key}</strong>: ${s.summary}
      </div>
    `)
    .join('');
}
```

### React Hook Example

```typescript
import { useState, useEffect } from 'react';

interface JiraSuggestion {
  key: string;
  summary: string;
}

function useJiraSuggestions(query: string, limit: number = 10) {
  const [suggestions, setSuggestions] = useState<JiraSuggestion[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (query.length < 2) {
      setSuggestions([]);
      return;
    }

    const controller = new AbortController();
    const fetchSuggestions = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await fetch(
          `http://localhost:8080/jira/suggest?q=${encodeURIComponent(query)}&limit=${limit}`,
          { signal: controller.signal }
        );

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        setSuggestions(data.suggestions);
      } catch (err: any) {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      } finally {
        setLoading(false);
      }
    };

    const debounce = setTimeout(fetchSuggestions, 300);

    return () => {
      clearTimeout(debounce);
      controller.abort();
    };
  }, [query, limit]);

  return { suggestions, loading, error };
}

// Usage in a component
function JiraIssueSelector() {
  const [query, setQuery] = useState('');
  const { suggestions, loading, error } = useJiraSuggestions(query);

  return (
    <div>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search Jira issues..."
      />
      {loading && <div>Loading...</div>}
      {error && <div>Error: {error}</div>}
      <ul>
        {suggestions.map((s) => (
          <li key={s.key}>
            <strong>{s.key}</strong>: {s.summary}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

## Search Behavior

The endpoint uses Jira Query Language (JQL) to search for issues:

- **Text Search**: Searches in issue summaries and descriptions
- **Key Search**: Matches issue keys (e.g., "PROJ" matches "PROJ-123", "PROJ-456")
- **Ordering**: Results are ordered by most recently updated
- **Wildcards**: Automatically adds wildcards for partial matching

Example JQL query generated:
```
text ~ "backend*" OR key ~ "backend*" ORDER BY updated DESC
```

## Response Status Codes

- **200 OK**: Suggestions returned successfully
- **400 Bad Request**: Invalid request (missing query, query too short)
- **405 Method Not Allowed**: Wrong HTTP method (only GET supported)
- **500 Internal Server Error**: Failed to query Jira API
- **503 Service Unavailable**: Jira integration not configured

## Error Handling

### Query Too Short
```json
Status: 400 Bad Request
Body: "query must be at least 2 characters"
```

### Missing Query Parameter
```json
Status: 400 Bad Request
Body: "query parameter 'q' is required"
```

### Jira Not Configured
```json
Status: 503 Service Unavailable
Body: "Jira integration not configured"
```

### Jira API Error
```json
Status: 500 Internal Server Error
Body: "failed to search Jira issues: jira API error: 401 - Unauthorized"
```

## Performance Considerations

- **Debouncing**: Implement debouncing on the client side (300-500ms) to avoid excessive API calls
- **Caching**: Consider caching results on the client for frequently searched terms
- **Limit**: Keep the limit reasonable (10-20) to ensure fast responses
- **Timeouts**: The Jira client has a 30-second timeout for API requests

## Security

- **API Tokens**: Store Jira API tokens securely, never commit them to version control
- **CORS**: Configure CORS properly if accessing from a web application
- **Authentication**: The endpoint doesn't require authentication to the Tock server, but does require valid Jira credentials configured on the server
- **Rate Limiting**: Be aware of Jira API rate limits (typically 1000 requests per hour for Cloud)

## Troubleshooting

### No results returned
1. Verify Jira credentials are correct
2. Check if the user has permission to view the issues
3. Ensure the query matches existing issue keys or summaries
4. Try increasing the limit parameter

### 503 Service Unavailable
The Jira integration is not configured. Check:
1. `~/.config/tock/tock.yaml` has jira configuration
2. Environment variables are set correctly
3. Server was restarted after configuration changes

### 500 Internal Server Error
Check the server logs for detailed error messages. Common issues:
- Invalid Jira URL
- Invalid credentials
- Network connectivity issues
- Jira API is down

### Slow responses
- Reduce the limit parameter
- Check network connectivity to Jira
- Consider implementing server-side caching for frequent queries

## Integration with Activities

When creating activities, use the suggested Jira issue key as an attribute:

```bash
# Start activity with Jira issue
curl -X POST http://localhost:8080/activity/start \
  -H "Content-Type: application/json" \
  -d '{
    "Description": "Fix authentication bug",
    "Project": "backend-api",
    "Attributes": [
      {
        "Key": "jira",
        "Value": "BACKEND-123"
      }
    ]
  }'
```

## Future Enhancements

Potential improvements for consideration:
- Server-side caching of search results
- Recent/favorite issues for faster selection
- Project-specific filtering
- Issue status filtering (only open issues)
- Assignee filtering
- Custom field support
