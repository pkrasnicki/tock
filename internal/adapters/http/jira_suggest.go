package http

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// JiraSuggestion represents a suggested Jira issue
type JiraSuggestion struct {
	Key     string `json:"key"`
	Summary string `json:"summary"`
}

// JiraSuggestResponse represents the response for Jira suggestions
type JiraSuggestResponse struct {
	Suggestions []JiraSuggestion `json:"suggestions"`
	Query       string           `json:"query"`
}

// JiraSuggest handles autosuggest requests for Jira issues
func (h *Handler) JiraSuggest(w http.ResponseWriter, r *http.Request) {
	if h.jiraClient == nil {
		http.Error(w, "Jira integration not configured", http.StatusServiceUnavailable)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Require at least 2 characters
	if len(query) < 2 {
		http.Error(w, "query must be at least 2 characters", http.StatusBadRequest)
		return
	}

	// Get optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Search Jira issues
	issues, err := h.jiraClient.SearchIssues(query, limit)
	if err != nil {
		http.Error(w, "failed to search Jira issues: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to suggestions
	suggestions := make([]JiraSuggestion, len(issues))
	for i, issue := range issues {
		suggestions[i] = JiraSuggestion{
			Key:     issue.Key,
			Summary: issue.Fields.Summary,
		}
	}

	response := JiraSuggestResponse{
		Suggestions: suggestions,
		Query:       query,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
