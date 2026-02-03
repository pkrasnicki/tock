package file

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/tock/internal/core/models"
)

var ErrSkip = errors.New("skip line")

const (
	timeLayoutMin = "2006-01-02 15:04"
	timeLayoutSec = "2006-01-02 15:04:05"
)

// splitFields splits a line by | separator, respecting quoted fields
func splitFields(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for i, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		case '"':
			inQuotes = !inQuotes
		case '|':
			if inQuotes {
				current.WriteRune(r)
			} else {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}

		// End of string
		if i == len(line)-1 {
			fields = append(fields, current.String())
		}
	}

	return fields
}

// unquoteField removes quotes and unescapes a field
func unquoteField(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
		// Unescape quotes
		s = strings.ReplaceAll(s, `\"`, `"`)
		s = strings.ReplaceAll(s, `\\`, `\`)
	}
	return s
}

// quoteField quotes a field if it contains special characters
func quoteField(s string) string {
	if strings.ContainsAny(s, "|\"\n") {
		// Escape backslashes and quotes
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return `"` + s + `"`
	}
	return s
}

func ParseActivity(line string) (*models.Activity, error) {
	parts := splitFields(line)
	if len(parts) < 3 {
		return nil, ErrSkip // Not an activity line
	}

	// Format: [ID|]time|project|description[|attributes] or time|project|description (legacy)
	var id, timePart, project, description string
	var attributes []models.Attribute

	if len(parts) >= 4 {
		// New format with ID
		id = strings.TrimSpace(unquoteField(parts[0]))
		timePart = strings.TrimSpace(unquoteField(parts[1]))
		project = unquoteField(parts[2])
		description = unquoteField(parts[3])

		// Check for attributes in 5th part
		if len(parts) >= 5 {
			attrPart := strings.TrimSpace(unquoteField(parts[4]))
			attributes = parseAttributes(attrPart)
		}
	} else {
		// Legacy format without ID
		id = "" // Will be generated if needed
		timePart = strings.TrimSpace(unquoteField(parts[0]))
		project = unquoteField(parts[1])
		description = unquoteField(parts[2])
	}

	// Ensure attributes is never nil
	if attributes == nil {
		attributes = []models.Attribute{}
	}

	var start, end time.Time
	var err error

	if strings.Contains(timePart, " - ") {
		times := strings.Split(timePart, " - ")
		start, err = parseTime(times[0])
		if err != nil {
			return nil, errors.Wrap(err, "parse start time")
		}
		end, err = parseTime(times[1])
		if err != nil {
			return nil, errors.Wrap(err, "parse end time")
		}

		// Generate stable ID if not present (legacy format)
		if id == "" {
			id = models.GenerateStableID(start, project, description)
		}

		return &models.Activity{
			ID:          id,
			StartTime:   start,
			EndTime:     &end,
			Project:     project,
			Description: description,
			Attributes:  attributes,
		}, nil
	}

	start, err = parseTime(timePart)
	if err != nil {
		return nil, errors.Wrap(err, "parse start time")
	}

	// Generate stable ID if not present (legacy format)
	if id == "" {
		id = models.GenerateStableID(start, project, description)
	}

	return &models.Activity{
		ID:          id,
		StartTime:   start,
		EndTime:     nil,
		Project:     project,
		Description: description,
		Attributes:  attributes,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	t, err := time.ParseInLocation(timeLayoutMin, s, time.Local)
	if err == nil {
		return t, nil
	}
	return time.ParseInLocation(timeLayoutSec, s, time.Local)
}

func FormatActivity(a models.Activity) string {
	startStr := a.StartTime.Format(timeLayoutSec)
	attributesStr := formatAttributes(a.Attributes)

	// Quote project and description if they contain pipe characters
	project := quoteField(a.Project)
	description := quoteField(a.Description)

	if a.EndTime != nil {
		endStr := a.EndTime.Format(timeLayoutSec)
		if attributesStr != "" {
			return fmt.Sprintf("%s | %s - %s | %s | %s | %s", a.ID, startStr, endStr, project, description, attributesStr)
		}
		return fmt.Sprintf("%s | %s - %s | %s | %s", a.ID, startStr, endStr, project, description)
	}
	if attributesStr != "" {
		return fmt.Sprintf("%s | %s | %s | %s | %s", a.ID, startStr, project, description, attributesStr)
	}
	return fmt.Sprintf("%s | %s | %s | %s", a.ID, startStr, project, description)
}

// parseAttributes parses comma-separated key=value pairs
func parseAttributes(s string) []models.Attribute {
	if s == "" {
		return []models.Attribute{}
	}

	var attrs []models.Attribute
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			attrs = append(attrs, models.Attribute{
				Key:   strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}
	}
	return attrs
}

// formatAttributes formats attributes as comma-separated key=value pairs
func formatAttributes(attrs []models.Attribute) string {
	if len(attrs) == 0 {
		return ""
	}

	var pairs []string
	for _, attr := range attrs {
		pairs = append(pairs, fmt.Sprintf("%s=%s", attr.Key, attr.Value))
	}
	return strings.Join(pairs, ",")
}
