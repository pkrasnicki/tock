package http

import (
	"net/http"
	"time"

	"github.com/kriuchkov/tock/internal/core/ports"
)

type Handler struct {
	service     ports.ActivityResolver
	corsConfig  CORSConfig
	broadcaster *EventBroadcaster
}

func NewHandler(service ports.ActivityResolver) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  DefaultCORSConfig(),
		broadcaster: NewEventBroadcaster(60 * time.Second),
	}
}

// NewHandlerWithCORS creates a new handler with custom CORS configuration
func NewHandlerWithCORS(service ports.ActivityResolver, corsConfig CORSConfig) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  corsConfig,
		broadcaster: NewEventBroadcaster(60 * time.Second),
	}
}

// NewHandlerWithEvents creates a new handler with events enabled
func NewHandlerWithEvents(service ports.ActivityResolver, broadcaster *EventBroadcaster) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  DefaultCORSConfig(),
		broadcaster: broadcaster,
	}
}

// NewHandlerWithCORSAndEvents creates a new handler with custom CORS and events
func NewHandlerWithCORSAndEvents(service ports.ActivityResolver, corsConfig CORSConfig, broadcaster *EventBroadcaster) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  corsConfig,
		broadcaster: broadcaster,
	}
}

// RegisterRoutes registers all activity routes with the default ServeMux
func (h *Handler) RegisterRoutes() {
	cors := CORSMiddleware(h.corsConfig)

	http.HandleFunc("/activity/start", cors(h.Start))
	http.HandleFunc("/activity/stop", cors(h.Stop))
	http.HandleFunc("/activity/add", cors(h.Add))
	http.HandleFunc("/activity/remove", cors(h.Remove))
	http.HandleFunc("/activity/list", cors(h.List))
	http.HandleFunc("/activity/current", cors(h.Current))
	http.HandleFunc("/activity/recent", cors(h.Recent))
	http.HandleFunc("/activity/report", cors(h.Report))
	http.HandleFunc("/activity/events", cors(h.Events))
}
