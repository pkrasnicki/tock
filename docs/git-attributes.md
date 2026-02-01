# Git Branch Attribute Mapping

Tock can automatically add attributes to activities based on the current git branch name. This is useful for tracking work context, billability, project types, and more.

## Configuration

Add `attribute_patterns` to your `~/.config/tock/tock.yaml`:

```yaml
attribute_patterns:
  - pattern: "^feature/(PROJ-\\d+)"
    attributes:
      type: "feature"
      ticket: "$1"           # Captures PROJ-123 from feature/PROJ-123
      billable: "true"
  
  - pattern: "^bugfix/.*"
    attributes:
      type: "bugfix"
      priority: "high"
  
  - pattern: "^([A-Z]+)-(\\d+)-.*"
    attributes:
      project: "$1"          # First capture group
      ticket-id: "$2"        # Second capture group
      full-ref: "$0"         # Full match
```

## Pattern Matching

- **pattern**: A regular expression to match against the branch name
- **attributes**: Key-value pairs to add when the pattern matches
- **Capture groups**: Use `$0` for full match, `$1`, `$2`, etc. for captured groups

### Capture Group Examples

```yaml
# Extract ticket number from branch name
- pattern: "feature/(PROJ-\\d+)"
  attributes:
    ticket: "$1"              # PROJ-242 from feature/PROJ-242

# Extract project and number separately  
- pattern: "^([A-Z]+)-(\\d+)"
  attributes:
    project: "$1"             # PROJ from PROJ-242
    number: "$2"              # 242 from PROJ-242

# Use full match
- pattern: "(feature|bugfix)/.*"
  attributes:
    branch-type: "$1"         # feature or bugfix
    full-branch: "$0"         # entire branch name
```

### Pattern Examples

```yaml
# Match feature branches
pattern: "^feature/.*"

# Match Jira-style ticket branches (PROJ-123, TICKET-456)
pattern: "^[A-Z]+-\\d+"

# Match any branch with "urgent" in the name
pattern: ".*urgent.*"

# Match specific branch names
pattern: "^(main|master|develop)$"
```

## How It Works

When you run `tock start`, the command:

1. Checks if you're in a git repository
2. Gets the current branch name
3. Matches it against all configured patterns
4. Adds all matching attributes to the activity

## Example

Given this configuration:

```yaml
attribute_patterns:
  - pattern: "^feature/(ACME-\\d+)"
    attributes:
      client: "acme-corp"
      ticket: "$1"
      billable: "true"
```

When you start an activity on branch `feature/ACME-123`:

```bash
$ tock start -p backend -d "Implementing API"
Started activity: backend | Implementing API at 14:30
Attributes: client=acme-corp, ticket=ACME-123, billable=true
```

## Multiple Matches

If a branch matches multiple patterns, all attributes from all matching patterns are added to the activity.

## Use Cases

- **Billability tracking**: Auto-mark client work as billable
- **Project categorization**: Tag work by project or team
- **Priority tracking**: Mark urgent/high-priority work
- **Compliance**: Track work that requires audit trails
- **Time allocation**: Categorize work for reporting
