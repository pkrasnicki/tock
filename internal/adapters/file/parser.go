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

func ParseActivity(line string) (*models.Activity, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 3 {
		return nil, ErrSkip // Not an activity line
	}

	// Format: [ID|]time|project|description[|key1=value1,key2=value2] or time|project|description (legacy)
	var id, timePart, project, description string
	var attributes []models.Attribute

	if len(parts) >= 4 {
		// New format with ID
		id = strings.TrimSpace(parts[0])
		timePart = strings.TrimSpace(parts[1])
		project = strings.TrimSpace(parts[2])
		description = strings.TrimSpace(parts[3])

		// Check for attributes in 5th part
		if len(parts) >= 5 {
			attributes = parseAttributes(strings.TrimSpace(parts[4]))
		}
	} else {
		// Legacy format without ID
		id = "" // Will be generated if needed
		timePart = strings.TrimSpace(parts[0])
		project = strings.TrimSpace(parts[1])
		description = strings.TrimSpace(parts[2])
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
	startStr := a.StartTime.Format(timeLayoutMin)
	attributesStr := formatAttributes(a.Attributes)

	if a.EndTime != nil {
		endStr := a.EndTime.Format(timeLayoutMin)
		if attributesStr != "" {
			return fmt.Sprintf("%s | %s - %s | %s | %s | %s", a.ID, startStr, endStr, a.Project, a.Description, attributesStr)
		}
		return fmt.Sprintf("%s | %s - %s | %s | %s", a.ID, startStr, endStr, a.Project, a.Description)
	}
	if attributesStr != "" {
		return fmt.Sprintf("%s | %s | %s | %s | %s", a.ID, startStr, a.Project, a.Description, attributesStr)
	}
	return fmt.Sprintf("%s | %s | %s | %s", a.ID, startStr, a.Project, a.Description)
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
