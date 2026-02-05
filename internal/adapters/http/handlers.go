package http

import (
	"net/http"
	"time"

	"github.com/kriuchkov/tock/internal/core/ports"
	"github.com/kriuchkov/tock/internal/jira"
)

type Handler struct {
	service     ports.ActivityResolver
	corsConfig  CORSConfig
	broadcaster *EventBroadcaster
	jiraClient  *jira.Client
	verbose     bool
}

func NewHandler(service ports.ActivityResolver) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  DefaultCORSConfig(),
		broadcaster: NewEventBroadcaster(60 * time.Second),
		jiraClient:  nil,
	}
}

// NewHandlerWithCORS creates a new handler with custom CORS configuration
func NewHandlerWithCORS(service ports.ActivityResolver, corsConfig CORSConfig) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  corsConfig,
		broadcaster: NewEventBroadcaster(60 * time.Second),
		jiraClient:  nil,
	}
}

// NewHandlerWithEvents creates a new handler with events enabled
func NewHandlerWithEvents(service ports.ActivityResolver, broadcaster *EventBroadcaster) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  DefaultCORSConfig(),
		broadcaster: broadcaster,
		jiraClient:  nil,
	}
}

// NewHandlerWithCORSAndEvents creates a new handler with custom CORS and events
func NewHandlerWithCORSAndEvents(service ports.ActivityResolver, corsConfig CORSConfig, broadcaster *EventBroadcaster) *Handler {
	return &Handler{
		service:     service,
		corsConfig:  corsConfig,
		broadcaster: broadcaster,
		jiraClient:  nil,
	}
}

// HandlerOptions configures Handler creation
type HandlerOptions struct {
	Service     ports.ActivityResolver
	CORSConfig  *CORSConfig
	Broadcaster *EventBroadcaster
	JiraClient  *jira.Client
	Verbose     bool
}

// NewHandlerWithOptions creates a new handler with all optional configurations
func NewHandlerWithOptions(opts HandlerOptions) *Handler {
	cors := DefaultCORSConfig()
	if opts.CORSConfig != nil {
		cors = *opts.CORSConfig
	}

	broadcaster := opts.Broadcaster
	if broadcaster == nil {
		broadcaster = NewEventBroadcaster(60 * time.Second)
	}

	return &Handler{
		service:     opts.Service,
		corsConfig:  cors,
		broadcaster: broadcaster,
		jiraClient:  opts.JiraClient,
		verbose:     opts.Verbose,
	}
}

// RegisterRoutes registers all activity routes with the default ServeMux
func (h *Handler) RegisterRoutes() {
	cors := CORSMiddleware(h.corsConfig)
	logging := LoggingMiddleware(h.verbose)

	// Chain middleware: logging -> cors -> handler
	http.HandleFunc("/activity/start", logging(cors(h.Start)))
	http.HandleFunc("/activity/stop", logging(cors(h.Stop)))
	http.HandleFunc("/activity/add", logging(cors(h.Add)))
	http.HandleFunc("/activity/remove", logging(cors(h.Remove)))
	http.HandleFunc("/activity/list", logging(cors(h.List)))
	http.HandleFunc("/activity/current", logging(cors(h.Current)))
	http.HandleFunc("/activity/recent", logging(cors(h.Recent)))
	http.HandleFunc("/activity/report", logging(cors(h.Report)))
	http.HandleFunc("/activity/events", logging(cors(h.Events)))
	http.HandleFunc("/jira/suggest", logging(cors(h.JiraSuggest)))
}
