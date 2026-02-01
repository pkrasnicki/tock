package http

import (
	"net/http"

	"github.com/kriuchkov/tock/internal/core/ports"
)

type Handler struct {
	service ports.ActivityResolver
}

func NewHandler(service ports.ActivityResolver) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all activity routes with the default ServeMux
func (h *Handler) RegisterRoutes() {
	http.HandleFunc("/activity/start", h.Start)
	http.HandleFunc("/activity/stop", h.Stop)
	http.HandleFunc("/activity/add", h.Add)
	http.HandleFunc("/activity/list", h.List)
	http.HandleFunc("/activity/current", h.Current)
	http.HandleFunc("/activity/recent", h.Recent)
	http.HandleFunc("/activity/report", h.Report)
}
