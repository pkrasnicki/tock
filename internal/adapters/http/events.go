package http

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

// Events handles long-polling requests for activity events
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	if h.broadcaster == nil {
		http.Error(w, "events not enabled", http.StatusServiceUnavailable)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate unique subscriber ID
	subscriberID := generateSubscriberID()

	// Subscribe to events
	subscriber := h.broadcaster.Subscribe(r.Context(), subscriberID)
	defer h.broadcaster.Unsubscribe(subscriberID)

	// Set headers for long-polling
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a timeout for the long-polling request
	timeout := time.After(h.broadcaster.GetTimeout())

	select {
	case event := <-subscriber.Events:
		// Event received, send it to the client
		if err := json.NewEncoder(w).Encode(event); err != nil {
			http.Error(w, "failed to encode event: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case <-timeout:
		// Timeout reached, send empty response to allow client to reconnect
		w.WriteHeader(http.StatusNoContent)

	case <-r.Context().Done():
		// Client disconnected
		w.WriteHeader(http.StatusRequestTimeout)
	}
}

// generateSubscriberID generates a random unique ID for a subscriber
func generateSubscriberID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(bytes)
}
