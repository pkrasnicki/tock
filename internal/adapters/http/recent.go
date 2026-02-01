package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kriuchkov/tock/internal/core/models"
)

func (h *Handler) Recent(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	activities, err := h.service.GetRecent(r.Context(), limit)
	if err != nil {
		http.Error(w, "failed to get recent activities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null when there are no activities
	if activities == nil {
		activities = []models.Activity{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}
