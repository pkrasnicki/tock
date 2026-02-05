package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kriuchkov/tock/internal/core/dto"
)

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	var req dto.StartActivityRequest

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		req.Description = r.URL.Query().Get("description")
		req.Project = r.URL.Query().Get("project")
	}

	if req.StartTime.IsZero() {
		req.StartTime = time.Now()
	}

	activity, err := h.service.Start(r.Context(), req)
	if err != nil {
		http.Error(w, "failed to start activity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the event if broadcaster is enabled
	if h.broadcaster != nil {
		h.broadcaster.BroadcastActivityStarted(activity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}
