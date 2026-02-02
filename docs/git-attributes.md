# Git Branch Pattern Matching

Tock can automatically extract project name, description, and attributes from the current git branch name and remote URLs. This eliminates the need to manually specify these values when starting activities.

## Configuration

Add `attribute_patterns` to your `~/.config/tock/tock.yaml`:

```yaml
attribute_patterns:
  # Extract ticket from branch and set project from remote
  - pattern: "^feature/(PROJ-\\d+)"
    project: "$remote_origin"           # Use mapped remote as project
    description: "Working on $1"        # Use capture group
    attributes:
      type: "feature"
      ticket: "$1"                      # Captures PROJ-123 from feature/PROJ-123
      billable: "true"
    remote_mappings:
      "github.com/company/backend": "Backend"
      "github.com/company/frontend": "Frontend"
  
  # Simple bugfix pattern
  - pattern: "^bugfix/(.*)"
    description: "Fix: $1"
    attributes:
      type: "bugfix"
      priority: "high"
  
  # Extract project and ticket from branch
  - pattern: "^([A-Z]+)-(\\d+)-.*"
    description: "$1-$2"
    attributes:
      project: "$1"                     # First capture group
      ticket-id: "$2"                   # Second capture group
```

## Pattern Configuration

Each pattern supports:

- **pattern**: Regular expression to match against the branch name
- **project**: Template for project name (optional, uses flags if not set)
- **description**: Template for activity description (optional, uses flags if not set)  
- **attributes**: Key-value pairs to add when the pattern matches
- **remote_mappings**: Map remote URLs to custom project names

## Variable Expansion

### Capture Groups

Use `$0` for full match, `$1`, `$2`, etc. for captured groups:

```yaml
# Extract ticket number from branch name
- pattern: "feature/(PROJ-\\d+)"
  description: "Working on $1"          # "Working on PROJ-242"
  attributes:
    ticket: "$1"                        # "PROJ-242"

# Extract project and number separately  
- pattern: "^([A-Z]+)-(\\d+)"
  description: "$1-$2"                  # "PROJ-242"
  attributes:
    project: "$1"                       # "PROJ"
    number: "$2"                        # "242"
```

### Git Remote Variables

Use `$remote_<name>` to reference git remote URLs:

```yaml
- pattern: ".*"                         # Match any branch
  project: "$remote_origin"
  attributes:
    remote: "$remote_upstream"
```

Remote variables are automatically populated from `git remote -v`:
- `$remote_origin` - origin remote URL
- `$remote_upstream` - upstream remote URL  
- `$remote_<any>` - any configured remote

### Remote Mappings

Map remote URLs to friendly names:

```yaml
- pattern: ".*"
  project: "$remote_origin"
  remote_mappings:
    # Exact URL match
    "git@github.com:company/backend.git": "Backend"
    "git@github.com:company/frontend.git": "Frontend"
    
    # Regex pattern match
    ".*github\\.com/company/.*": "CompanyProjects"
    ".*gitlab\\.com/personal/.*": "PersonalProjects"
```

## How It Works

When you run `tock start`:

1. Checks if you're in a git repository
2. Gets the current branch name and remote URLs
3. Matches branch against all configured patterns
4. Extracts project, description, and attributes from first matching pattern
5. Uses flag values if no pattern match or pattern doesn't specify project/description

## Example Workflows

### Automatic Project from Remote

```yaml
attribute_patterns:
  - pattern: ".*"
    project: "$remote_origin"
    remote_mappings:
      "github.com/acme/api": "ACME-API"
      "github.com/acme/web": "ACME-Web"
```

```bash
# In repository github.com/acme/api on branch feature/auth
$ tock start -d "Implementing OAuth"
Started activity: ACME-API | Implementing OAuth at 14:30
```

### Auto-populate from Branch Pattern

```yaml
attribute_patterns:
  - pattern: "^feature/(JIRA-\\d+)-(.*)"
    project: "MyProject"
    description: "$1: $2"
    attributes:
      ticket: "$1"
      type: "feature"
```

```bash
# On branch feature/JIRA-123-add-login
$ tock start
Started activity: MyProject | JIRA-123: add-login at 14:30
Attributes: ticket=JIRA-123, type=feature
```

### Mix Patterns and Flags

```yaml
attribute_patterns:
  - pattern: "^feature/(PROJ-\\d+)"
    attributes:
      ticket: "$1"
```

```bash
# Pattern provides ticket, flags provide project and description
$ tock start -p Backend -d "API work"
Started activity: Backend | API work at 14:30
Attributes: ticket=PROJ-242
```

## Pattern Examples

```yaml
# Match feature branches
pattern: "^feature/.*"

# Match Jira-style ticket branches (PROJ-123, TICKET-456)
pattern: "^[A-Z]+-\\d+"

# Match any branch with "urgent" in the name
pattern: ".*urgent.*"

# Match specific branch names
pattern: "^(main|master|develop)$"

# Extract ticket and description
pattern: "^feature/(\\w+-\\d+)-(.*)"
```
