# Jira Synchronization

Tock can automatically synchronize your time tracking activities with Jira worklogs. This keeps your local time tracking in sync with Jira's time tracking system.

## Configuration

Add Jira credentials to your `~/.config/tock/tock.yaml`:

```yaml
jira:
  url: "https://your-domain.atlassian.net"
  username: "your-email@example.com"
  api_token: "your-api-token"
```

### Getting a Jira API Token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a name (e.g., "Tock Sync")
4. Copy the token and add it to your config

You can also use environment variables:
```bash
export TOCK_JIRA_URL="https://your-domain.atlassian.net"
export TOCK_JIRA_USERNAME="your-email@example.com"
export TOCK_JIRA_API_TOKEN="your-api-token"
```

## Usage

### Mark Activities for Sync

Add a `jira` attribute to activities you want to sync with Jira:

```bash
# Manual start with jira attribute
tock start -p backend -d "API implementation"
# Then add jira attribute manually to the file, or...

# Use git branch patterns to auto-add jira attribute
# Configure in tock.yaml:
attribute_patterns:
  - pattern: "^feature/(PROJ-\\d+)"
    attributes:
      jira: "$1"
```

With this pattern, if you're on branch `feature/PROJ-123`, the activity will automatically get `jira=PROJ-123`.

### Sync Command

Run the sync command to synchronize all activities:

```bash
tock sync
```

The command will:
- ✓ Add new worklogs for unsynced activities with jira attribute
- ✓ Update existing worklogs if activity times changed
- ✓ Delete worklogs if jira attribute was removed
- ✓ Move worklogs if jira attribute value changed to a different issue

## How It Works

### Activity States

**Not Synced**
- Activity has `jira` attribute
- No `JiraWorklogID` stored
- → Will be added to Jira on next sync

**Synced**
- Activity has `jira` attribute
- Has `JiraWorklogID` stored
- → Will be updated if times change

**Removed from Sync**
- Activity no longer has `jira` attribute
- Has `JiraWorklogID` stored
- → Will be deleted from Jira on next sync

**Skipped**
- Activity has no `jira` attribute
- No `JiraWorklogID` stored
- → Ignored during sync

### Sync Metadata

Tock stores sync metadata in the activity file:

```
ID | time | project | description | attributes | jiraWorklogID:jiraIssueKey
```

Example:
```
abc123 | 2026-02-02 09:00:00 - 2026-02-02 11:00:00 | backend | API work | jira=PROJ-123 | 10001:PROJ-123
```

## Scenarios

### 1. New Activity with Jira Attribute

```bash
tock start -p backend -d "Fixing bug"
# Add jira=PROJ-456 attribute
tock stop
tock sync
```
→ Creates worklog in PROJ-456

### 2. Update Activity Time

```bash
# Edit activity in file to adjust times
tock sync
```
→ Updates existing worklog in Jira

### 3. Remove Jira Attribute

```bash
# Remove jira attribute from activity
tock sync
```
→ Deletes worklog from Jira

### 4. Change Jira Issue

```bash
# Change jira=PROJ-123 to jira=PROJ-456
tock sync
```
→ Deletes worklog from PROJ-123, creates new worklog in PROJ-456

### 5. Running Activities

Running activities (without end time) are automatically skipped during sync.

## Sync Output

```bash
$ tock sync
Synchronizing activities with Jira...

✓ Sync completed:
  • Synced (new):    5
  • Updated:         2
  • Deleted:         1
  • Skipped:         10

✗ Errors:
  • Failed to add worklog to PROJ-999: Issue not found
```

## Best Practices

1. **Use git branch patterns** to automatically add jira attributes
2. **Sync regularly** to keep Jira up to date
3. **Review before syncing** - check which activities have jira attribute
4. **Keep API token secure** - use environment variables or restrict file permissions

## Troubleshooting

### "Jira configuration is missing"
Add jira configuration to `~/.config/tock/tock.yaml` or set environment variables.

### "Issue not found"
The Jira issue key in the `jira` attribute doesn't exist or you don't have access.

### "Failed to add worklog"
Check:
- API token is valid
- You have permission to log work on the issue
- The Jira URL is correct
- Network connectivity

### Activities not syncing
Verify:
- Activity has `jira` attribute with valid issue key
- Activity is completed (has end time)
- Run `tock sync` to trigger synchronization

## Integration with Git Patterns

Combine with attribute patterns for automatic workflow:

```yaml
attribute_patterns:
  - pattern: "^feature/([A-Z]+-\\d+)"
    attributes:
      jira: "$1"
      type: "feature"
  - pattern: "^bugfix/([A-Z]+-\\d+)"
    attributes:
      jira: "$1"
      type: "bugfix"
```

Now:
1. Work on branch `feature/PROJ-123`
2. `tock start -p backend -d "Implementation"` → auto-adds `jira=PROJ-123`
3. `tock stop`
4. `tock sync` → worklog appears in PROJ-123

## API Endpoints Used

- `POST /rest/api/3/issue/{issueKey}/worklog` - Add worklog
- `PUT /rest/api/3/issue/{issueKey}/worklog/{id}` - Update worklog
- `DELETE /rest/api/3/issue/{issueKey}/worklog/{id}` - Delete worklog
