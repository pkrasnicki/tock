package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Activity represents a time tracking entry.
type Activity struct {
	ID          string      `json:"ID"`
	Description string      `json:"Description"`
	Project     string      `json:"Project"`
	StartTime   time.Time   `json:"StartTime"`
	EndTime     *time.Time  `json:"EndTime,omitempty"`
	Attributes  []Attribute `json:"Attributes"`
}

// NewActivity creates a new Activity with an empty attributes slice
func NewActivity() Activity {
	return Activity{
		Attributes: []Attribute{},
	}
}

// GenerateID creates a unique identifier for an activity
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateStableID creates a deterministic ID based on activity properties
// This is used for legacy activities that don't have IDs yet
func GenerateStableID(startTime time.Time, project, description string) string {
	// Create a stable hash from the activity's immutable properties
	data := fmt.Sprintf("%d|%s|%s", startTime.Unix(), project, description)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes (32 hex chars)
}

// Duration returns the duration of the activity.
// If the activity is active, it returns the duration from StartTime to now.
func (a Activity) Duration() time.Duration {
	if a.EndTime != nil {
		return a.EndTime.Sub(a.StartTime)
	}
	return time.Since(a.StartTime)
}
