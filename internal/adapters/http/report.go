package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kriuchkov/tock/internal/core/dto"
)

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
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

	report, err := h.service.GetReport(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to get report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
