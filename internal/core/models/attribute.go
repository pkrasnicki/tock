package models

// Attribute represents a key-value pair that can be attached to an activity
type Attribute struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}
