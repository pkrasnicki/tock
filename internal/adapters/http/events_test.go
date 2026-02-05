package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kriuchkov/tock/internal/core/models"
)

func TestEventBroadcaster_Subscribe(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	ctx := context.Background()

	subscriber := broadcaster.Subscribe(ctx, "test-1")

	if subscriber == nil {
		t.Fatal("expected subscriber to be created")
	}

	if subscriber.ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", subscriber.ID)
	}

	if subscriber.Events == nil {
		t.Fatal("expected events channel to be created")
	}
}

func TestEventBroadcaster_Broadcast(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	ctx := context.Background()

	subscriber := broadcaster.Subscribe(ctx, "test-1")

	activity := &models.Activity{
		ID:          "123",
		Description: "Test Activity",
		StartTime:   time.Now(),
	}

	// Broadcast in goroutine to prevent blocking
	go broadcaster.BroadcastActivityStarted(activity)

	// Wait for event with timeout
	select {
	case event := <-subscriber.Events:
		if event.Type != EventActivityStarted {
			t.Errorf("expected EventActivityStarted, got %s", event.Type)
		}
		if event.Activity == nil {
			t.Fatal("expected activity in event")
		}
		if event.Activity.ID != "123" {
			t.Errorf("expected activity ID 123, got %s", event.Activity.ID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBroadcaster_MultipleSubscribers(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	ctx := context.Background()

	sub1 := broadcaster.Subscribe(ctx, "test-1")
	sub2 := broadcaster.Subscribe(ctx, "test-2")

	activity := &models.Activity{
		ID:          "123",
		Description: "Test Activity",
		StartTime:   time.Now(),
	}

	go broadcaster.BroadcastActivityStarted(activity)

	// Both subscribers should receive the event
	received := 0
	timeout := time.After(1 * time.Second)

	for received < 2 {
		select {
		case <-sub1.Events:
			received++
		case <-sub2.Events:
			received++
		case <-timeout:
			t.Fatalf("timeout: only received %d/2 events", received)
		}
	}
}

func TestEventBroadcaster_Unsubscribe(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	ctx := context.Background()

	subscriber := broadcaster.Subscribe(ctx, "test-1")
	broadcaster.Unsubscribe("test-1")

	// Channel should be closed
	_, ok := <-subscriber.Events
	if ok {
		t.Error("expected channel to be closed after unsubscribe")
	}
}

func TestEvents_LongPollingTimeout(t *testing.T) {
	broadcaster := NewEventBroadcaster(1 * time.Second)
	handler := NewHandlerWithEvents(nil, broadcaster)

	req := httptest.NewRequest(http.MethodGet, "/activity/events", nil)
	w := httptest.NewRecorder()

	// Call Events handler
	handler.Events(w, req)

	// Should timeout and return 204
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestEvents_ReceiveEvent(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	handler := NewHandlerWithEvents(nil, broadcaster)

	req := httptest.NewRequest(http.MethodGet, "/activity/events", nil)
	w := httptest.NewRecorder()

	// Broadcast event after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		activity := &models.Activity{
			ID:          "123",
			Description: "Test Activity",
			StartTime:   time.Now(),
		}
		broadcaster.BroadcastActivityStarted(activity)
	}()

	// Call Events handler
	handler.Events(w, req)

	// Should receive event and return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Parse response
	var event ActivityEvent
	if err := json.Unmarshal(w.Body.Bytes(), &event); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if event.Type != EventActivityStarted {
		t.Errorf("expected EventActivityStarted, got %s", event.Type)
	}
}

func TestEvents_MethodNotAllowed(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	handler := NewHandlerWithEvents(nil, broadcaster)

	req := httptest.NewRequest(http.MethodPost, "/activity/events", nil)
	w := httptest.NewRecorder()

	handler.Events(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestEventTypes(t *testing.T) {
	broadcaster := NewEventBroadcaster(5 * time.Second)
	ctx := context.Background()
	subscriber := broadcaster.Subscribe(ctx, "test-1")

	activity := &models.Activity{
		ID:          "123",
		Description: "Test",
		StartTime:   time.Now(),
	}

	tests := []struct {
		name          string
		broadcast     func()
		expectedType  ActivityEventType
		expectActivity bool
	}{
		{
			name: "activity started",
			broadcast: func() {
				broadcaster.BroadcastActivityStarted(activity)
			},
			expectedType:  EventActivityStarted,
			expectActivity: true,
		},
		{
			name: "activity stopped",
			broadcast: func() {
				broadcaster.BroadcastActivityStopped(activity)
			},
			expectedType:  EventActivityStopped,
			expectActivity: true,
		},
		{
			name: "activity added",
			broadcast: func() {
				broadcaster.BroadcastActivityAdded(activity)
			},
			expectedType:  EventActivityAdded,
			expectActivity: true,
		},
		{
			name: "activity removed",
			broadcast: func() {
				broadcaster.BroadcastActivityRemoved("123")
			},
			expectedType:  EventActivityRemoved,
			expectActivity: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go tt.broadcast()

			select {
			case event := <-subscriber.Events:
				if event.Type != tt.expectedType {
					t.Errorf("expected %s, got %s", tt.expectedType, event.Type)
				}
				if tt.expectActivity && event.Activity == nil {
					t.Error("expected activity in event")
				}
				if !tt.expectActivity && event.Activity != nil {
					t.Error("expected no activity in event")
				}
			case <-time.After(1 * time.Second):
				t.Fatal("timeout waiting for event")
			}
		})
	}
}
