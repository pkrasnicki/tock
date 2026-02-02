package syncstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-faster/errors"
	"github.com/kriuchkov/tock/internal/core/models"
)

// Repository manages Jira sync state persistence
type Repository struct {
	filePath string
	mu       sync.RWMutex
}

// NewRepository creates a new sync state repository
func NewRepository(dataDir string) *Repository {
	syncFilePath := filepath.Join(dataDir, ".tock-jira-sync.json")
	return &Repository{
		filePath: syncFilePath,
	}
}

// Load loads all sync states from storage
func (r *Repository) Load() (map[string]*models.JiraSyncState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	states := make(map[string]*models.JiraSyncState)

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return states, nil // Empty map if file doesn't exist
		}
		return nil, errors.Wrap(err, "read sync state file")
	}

	if len(data) == 0 {
		return states, nil
	}

	var stateList []models.JiraSyncState
	if err := json.Unmarshal(data, &stateList); err != nil {
		return nil, errors.Wrap(err, "unmarshal sync states")
	}

	// Convert to map keyed by ActivityID
	for i := range stateList {
		states[stateList[i].ActivityID] = &stateList[i]
	}

	return states, nil
}

// Save saves all sync states to storage
func (r *Repository) Save(states map[string]*models.JiraSyncState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert map to slice
	var stateList []models.JiraSyncState
	for _, state := range states {
		stateList = append(stateList, *state)
	}

	data, err := json.MarshalIndent(stateList, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal sync states")
	}

	// Ensure directory exists
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return errors.Wrap(err, "create directory")
	}

	if err := os.WriteFile(r.filePath, data, 0640); err != nil {
		return errors.Wrap(err, "write sync state file")
	}

	return nil
}

// Get retrieves sync state for a specific activity
func (r *Repository) Get(activityID string) (*models.JiraSyncState, error) {
	states, err := r.Load()
	if err != nil {
		return nil, err
	}
	return states[activityID], nil
}

// Set stores sync state for a specific activity
func (r *Repository) Set(state *models.JiraSyncState) error {
	states, err := r.Load()
	if err != nil {
		return err
	}
	states[state.ActivityID] = state
	return r.Save(states)
}

// Delete removes sync state for a specific activity
func (r *Repository) Delete(activityID string) error {
	states, err := r.Load()
	if err != nil {
		return err
	}
	delete(states, activityID)
	return r.Save(states)
}
