package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kriuchkov/tock/internal/jira"
)

func TestJiraSuggest_NotConfigured(t *testing.T) {
	handler := NewHandlerWithOptions(HandlerOptions{
		Service:    nil,
		JiraClient: nil,
	})

	req := httptest.NewRequest(http.MethodGet, "/jira/suggest?q=test", nil)
	w := httptest.NewRecorder()

	handler.JiraSuggest(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestJiraSuggest_MethodNotAllowed(t *testing.T) {
	handler := NewHandlerWithOptions(HandlerOptions{
		Service:    nil,
		JiraClient: &jira.Client{},
	})

	req := httptest.NewRequest(http.MethodPost, "/jira/suggest?q=test", nil)
	w := httptest.NewRecorder()

	handler.JiraSuggest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestJiraSuggest_MissingQuery(t *testing.T) {
	handler := NewHandlerWithOptions(HandlerOptions{
		Service:    nil,
		JiraClient: &jira.Client{},
	})

	req := httptest.NewRequest(http.MethodGet, "/jira/suggest", nil)
	w := httptest.NewRecorder()

	handler.JiraSuggest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	if !contains(w.Body.String(), "required") {
		t.Errorf("expected error message about required parameter, got %s", w.Body.String())
	}
}

func TestJiraSuggest_QueryTooShort(t *testing.T) {
	handler := NewHandlerWithOptions(HandlerOptions{
		Service:    nil,
		JiraClient: &jira.Client{},
	})

	req := httptest.NewRequest(http.MethodGet, "/jira/suggest?q=a", nil)
	w := httptest.NewRecorder()

	handler.JiraSuggest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	if !contains(w.Body.String(), "at least 2 characters") {
		t.Errorf("expected error message about minimum length, got %s", w.Body.String())
	}
}

func TestJiraSuggest_ResponseStructure(t *testing.T) {
	// This test verifies the response structure is correct
	// In a real scenario, you would mock the Jira client

	response := JiraSuggestResponse{
		Query: "test",
		Suggestions: []JiraSuggestion{
			{
				Key:     "PROJ-123",
				Summary: "Test issue",
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded JiraSuggestResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Query != "test" {
		t.Errorf("expected query 'test', got %s", decoded.Query)
	}

	if len(decoded.Suggestions) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(decoded.Suggestions))
	}

	if decoded.Suggestions[0].Key != "PROJ-123" {
		t.Errorf("expected key 'PROJ-123', got %s", decoded.Suggestions[0].Key)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
