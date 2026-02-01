package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/kriuchkov/tock/internal/core/dto"
	"github.com/kriuchkov/tock/internal/core/models"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var filter dto.ActivityFilter

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.FromDate = &from
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ToDate = &to
		}
	}

	if project := r.URL.Query().Get("project"); project != "" {
		filter.Project = &project
	}

	if runningStr := r.URL.Query().Get("running"); runningStr != "" {
		if running, err := strconv.ParseBool(runningStr); err == nil {
			filter.IsRunning = &running
		}
	}

	activities, err := h.service.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to list activities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null when there are no activities
	if activities == nil {
		activities = []models.Activity{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}
