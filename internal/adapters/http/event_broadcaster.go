package http

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/kriuchkov/tock/internal/core/models"
)

// ActivityEventType represents the type of activity event
type ActivityEventType string

const (
	EventActivityStarted ActivityEventType = "activity_started"
	EventActivityStopped ActivityEventType = "activity_stopped"
	EventActivityAdded   ActivityEventType = "activity_added"
	EventActivityRemoved ActivityEventType = "activity_removed"
)

// ActivityEvent represents an event that occurred with an activity
type ActivityEvent struct {
	Type      ActivityEventType `json:"type"`
	Activity  *models.Activity  `json:"activity,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// EventSubscriber represents a client subscribed to events
type EventSubscriber struct {
	ID      string
	Events  chan ActivityEvent
	ctx     context.Context
	created time.Time
}

// EventBroadcaster manages event subscribers and broadcasts events
type EventBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[string]*EventSubscriber
	timeout     time.Duration
}

// NewEventBroadcaster creates a new event broadcaster
func NewEventBroadcaster(timeout time.Duration) *EventBroadcaster {
	if timeout == 0 {
		timeout = 60 * time.Second // Default 60 seconds timeout for long-polling
	}

	broadcaster := &EventBroadcaster{
		subscribers: make(map[string]*EventSubscriber),
		timeout:     timeout,
	}

	// Start cleanup goroutine to remove stale subscribers
	go broadcaster.cleanupStaleSubscribers()

	return broadcaster
}

// Subscribe adds a new subscriber and returns the event channel
func (b *EventBroadcaster) Subscribe(ctx context.Context, id string) *EventSubscriber {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If subscriber already exists, remove it first
	if existing, ok := b.subscribers[id]; ok {
		close(existing.Events)
		delete(b.subscribers, id)
	}

	subscriber := &EventSubscriber{
		ID:      id,
		Events:  make(chan ActivityEvent, 10), // Buffer to prevent blocking
		ctx:     ctx,
		created: time.Now(),
	}

	b.subscribers[id] = subscriber

	return subscriber
}

// Unsubscribe removes a subscriber
func (b *EventBroadcaster) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subscriber, ok := b.subscribers[id]; ok {
		close(subscriber.Events)
		delete(b.subscribers, id)
	}
}

// Broadcast sends an event to all subscribers
func (b *EventBroadcaster) Broadcast(event ActivityEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for id, subscriber := range b.subscribers {
		select {
		case subscriber.Events <- event:
			// Event sent successfully
		default:
			// Channel buffer is full, skip this subscriber
			// In production, you might want to log this
			_ = id
		}
	}
}

// BroadcastActivityStarted broadcasts an activity started event
func (b *EventBroadcaster) BroadcastActivityStarted(activity *models.Activity) {
	b.Broadcast(ActivityEvent{
		Type:      EventActivityStarted,
		Activity:  activity,
		Timestamp: time.Now(),
	})
}

// BroadcastActivityStopped broadcasts an activity stopped event
func (b *EventBroadcaster) BroadcastActivityStopped(activity *models.Activity) {
	b.Broadcast(ActivityEvent{
		Type:      EventActivityStopped,
		Activity:  activity,
		Timestamp: time.Now(),
	})
}

// BroadcastActivityAdded broadcasts an activity added event
func (b *EventBroadcaster) BroadcastActivityAdded(activity *models.Activity) {
	b.Broadcast(ActivityEvent{
		Type:      EventActivityAdded,
		Activity:  activity,
		Timestamp: time.Now(),
	})
}

// BroadcastActivityRemoved broadcasts an activity removed event
func (b *EventBroadcaster) BroadcastActivityRemoved(activityID string) {
	b.Broadcast(ActivityEvent{
		Type:      EventActivityRemoved,
		Timestamp: time.Now(),
	})
}

// GetTimeout returns the configured timeout for long-polling
func (b *EventBroadcaster) GetTimeout() time.Duration {
	return b.timeout
}

// cleanupStaleSubscribers periodically removes subscribers that have been waiting too long
func (b *EventBroadcaster) cleanupStaleSubscribers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for id, subscriber := range b.subscribers {
			// Check if subscriber context is done or if it's been around for too long
			select {
			case <-subscriber.ctx.Done():
				close(subscriber.Events)
				delete(b.subscribers, id)
			default:
				if now.Sub(subscriber.created) > b.timeout*2 {
					close(subscriber.Events)
					delete(b.subscribers, id)
				}
			}
		}
		b.mu.Unlock()
	}
}

// MarshalEventToJSON marshals an event to JSON
func MarshalEventToJSON(event ActivityEvent) ([]byte, error) {
	return json.Marshal(event)
}
