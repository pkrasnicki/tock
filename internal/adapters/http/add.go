package http

import (
	"encoding/json"
	"net/http"

	"github.com/kriuchkov/tock/internal/core/dto"
)

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var addReq dto.AddActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&addReq); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	activity, err := h.service.Add(r.Context(), addReq)
	if err != nil {
		http.Error(w, "failed to add activity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}
