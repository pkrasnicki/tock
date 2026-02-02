package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/tock/internal/adapters/syncstate"
	"github.com/kriuchkov/tock/internal/config"
	"github.com/kriuchkov/tock/internal/core/dto"
	"github.com/kriuchkov/tock/internal/core/models"
	"github.com/kriuchkov/tock/internal/core/ports"
	"github.com/kriuchkov/tock/internal/jira"
)

type Service struct {
	repo       ports.ActivityRepository
	syncRepo   *syncstate.Repository
	jiraClient *jira.Client
}

type SyncResult struct {
	Synced  int
	Updated int
	Deleted int
	Skipped int
	Errors  []string
}

func NewService(repo ports.ActivityRepository, syncRepo *syncstate.Repository, cfg config.JiraConfig) *Service {
	return &Service{
		repo:       repo,
		syncRepo:   syncRepo,
		jiraClient: jira.NewClient(cfg.URL, cfg.Username, cfg.APIToken),
	}
}

// Sync synchronizes activities with Jira
func (s *Service) Sync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{}

	// Load current sync states
	syncStates, err := s.syncRepo.Load()
	if err != nil {
		return nil, errors.Wrap(err, "load sync states")
	}

	// Get all activities (we'll filter out running ones)
	activities, err := s.repo.Find(ctx, dto.ActivityFilter{})
	if err != nil {
		return nil, errors.Wrap(err, "find activities")
	}

	// Track which activities we've seen
	seenActivityIDs := make(map[string]bool)

	for i := range activities {
		activity := &activities[i]
		seenActivityIDs[activity.ID] = true

		// Skip running activities
		if activity.EndTime == nil {
			result.Skipped++
			continue
		}

		// Find jira attribute
		jiraIssue := getJiraAttribute(activity.Attributes)
		syncState := syncStates[activity.ID]

		if jiraIssue == "" {
			// No jira attribute
			// If this activity was previously synced, delete from Jira
			if syncState != nil {
				if err := s.jiraClient.DeleteWorklog(syncState.IssueKey, syncState.WorklogID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete worklog %s from %s: %v", syncState.WorklogID, syncState.IssueKey, err))
				} else {
					result.Deleted++
					// Remove sync state
					if err := s.syncRepo.Delete(activity.ID); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete sync state: %v", err))
					}
				}
			} else {
				result.Skipped++
			}
			continue
		}

		// Activity has jira attribute
		if syncState != nil {
			// Already synced - check if we need to update or if issue changed
			if syncState.IssueKey != jiraIssue {
				// Issue key changed - delete from old, add to new
				if err := s.jiraClient.DeleteWorklog(syncState.IssueKey, syncState.WorklogID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete worklog from old issue %s: %v", syncState.IssueKey, err))
					continue
				}

				// Add to new issue
				if err := s.addWorklog(activity, jiraIssue); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to add worklog to new issue %s: %v", jiraIssue, err))
				} else {
					result.Updated++
				}
			} else {
				// Same issue - update worklog
				updated, err := s.updateWorklog(activity, syncState)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to update worklog %s: %v", syncState.WorklogID, err))
				} else if updated {
					result.Updated++
				} else {
					result.Skipped++
				}
			}
		} else {
			// Not synced yet - add new worklog
			if err := s.addWorklog(activity, jiraIssue); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to add worklog to %s: %v", jiraIssue, err))
			} else {
				result.Synced++
			}
		}
	}

	// Check for deleted activities (have sync state but activity no longer exists)
	for activityID, syncState := range syncStates {
		if !seenActivityIDs[activityID] {
			// Activity was deleted, remove worklog from Jira
			if err := s.jiraClient.DeleteWorklog(syncState.IssueKey, syncState.WorklogID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete worklog for deleted activity %s: %v", activityID, err))
			} else {
				result.Deleted++
				// Remove sync state
				if err := s.syncRepo.Delete(activityID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete sync state: %v", err))
				}
			}
		}
	}

	return result, nil
}

func (s *Service) addWorklog(activity *models.Activity, issueKey string) error {
	// Skip incomplete activities (no end time)
	if activity.EndTime == nil {
		return errors.New("activity has no end time")
	}

	duration := int(activity.Duration().Seconds())
	if duration < 60 {
		return errors.New("activity duration must be at least 60 seconds")
	}

	req := jira.WorklogRequest{
		Comment:          jira.NewComment(activity.Description),
		Started:          formatJiraTime(activity.StartTime),
		TimeSpentSeconds: duration,
	}

	resp, err := s.jiraClient.AddWorklog(issueKey, req)
	if err != nil {
		return err
	}

	// Store sync state with synced data
	syncState := &models.JiraSyncState{
		ActivityID:        activity.ID,
		WorklogID:         resp.ID,
		IssueKey:          issueKey,
		LastSynced:        time.Now().UTC().Format(time.RFC3339),
		SyncedStartTime:   activity.StartTime.UTC().Format(time.RFC3339),
		SyncedEndTime:     activity.EndTime.UTC().Format(time.RFC3339),
		SyncedDescription: activity.Description,
	}

	return s.syncRepo.Set(syncState)
}

func (s *Service) updateWorklog(activity *models.Activity, syncState *models.JiraSyncState) (bool, error) {
	// Skip incomplete activities (no end time)
	if activity.EndTime == nil {
		return false, errors.New("activity has no end time")
	}

	duration := int(activity.Duration().Seconds())
	if duration < 60 {
		return false, errors.New("activity duration must be at least 60 seconds")
	}

	// Check if data has changed since last sync
	currentStartTime := activity.StartTime.UTC().Format(time.RFC3339)
	currentEndTime := activity.EndTime.UTC().Format(time.RFC3339)
	currentDescription := activity.Description

	if syncState.SyncedStartTime == currentStartTime &&
		syncState.SyncedEndTime == currentEndTime &&
		syncState.SyncedDescription == currentDescription {
		// No changes detected, skip update
		return false, nil
	}

	req := jira.WorklogRequest{
		Comment:          jira.NewComment(activity.Description),
		Started:          formatJiraTime(activity.StartTime),
		TimeSpentSeconds: duration,
	}

	if err := s.jiraClient.UpdateWorklog(syncState.IssueKey, syncState.WorklogID, req); err != nil {
		return false, err
	}

	// Update sync state with new data
	syncState.LastSynced = time.Now().UTC().Format(time.RFC3339)
	syncState.SyncedStartTime = currentStartTime
	syncState.SyncedEndTime = currentEndTime
	syncState.SyncedDescription = currentDescription
	return true, s.syncRepo.Set(syncState)
}

// getJiraAttribute finds the value of the "jira" attribute
func getJiraAttribute(attributes []models.Attribute) string {
	for _, attr := range attributes {
		if attr.Key == "jira" {
			return attr.Value
		}
	}
	return ""
}

// formatJiraTime formats time for Jira API (2026-02-01T14:15:00.000+0000)
func formatJiraTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000-0700")
}
