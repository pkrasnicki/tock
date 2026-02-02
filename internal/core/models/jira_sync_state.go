package models

// JiraSyncState tracks the synchronization state between a Tock activity and a Jira worklog
type JiraSyncState struct {
	ActivityID        string `json:"ActivityID"`
	WorklogID         string `json:"WorklogID"`
	IssueKey          string `json:"IssueKey"`
	LastSynced        string `json:"LastSynced"`        // ISO 8601 timestamp
	SyncedStartTime   string `json:"SyncedStartTime"`   // ISO 8601 timestamp
	SyncedEndTime     string `json:"SyncedEndTime"`     // ISO 8601 timestamp
	SyncedDescription string `json:"SyncedDescription"` // Synced activity description
}
