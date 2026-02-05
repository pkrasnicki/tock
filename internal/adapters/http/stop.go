package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kriuchkov/tock/internal/core/dto"
)

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	var stopReq dto.StopActivityRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&stopReq); err != nil {
			stopReq.EndTime = time.Now()
		}
	} else {
		stopReq.EndTime = time.Now()
	}

	if stopReq.EndTime.IsZero() {
		stopReq.EndTime = time.Now()
	}

	activity, err := h.service.Stop(r.Context(), stopReq)
	if err != nil {
		http.Error(w, "failed to stop activity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the event if broadcaster is enabled
	if h.broadcaster != nil {
		h.broadcaster.BroadcastActivityStopped(activity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}
