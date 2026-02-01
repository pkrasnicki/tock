package http

import (
	"encoding/json"
	"net/http"

	"github.com/kriuchkov/tock/internal/core/dto"
)

func (h *Handler) Current(w http.ResponseWriter, r *http.Request) {
	isRunning := true
	filter := dto.ActivityFilter{
		IsRunning: &isRunning,
	}

	activities, err := h.service.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to get current activities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(activities) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}
