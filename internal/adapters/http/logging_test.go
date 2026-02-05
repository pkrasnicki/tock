package http

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware_Disabled(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(log.Writer())

	handler := LoggingMiddleware(false)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// Should not log anything when verbose is false
	if logBuffer.Len() > 0 {
		t.Errorf("expected no logs when verbose=false, got: %s", logBuffer.String())
	}
}

func TestLoggingMiddleware_Enabled(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(log.Writer())

	handler := LoggingMiddleware(true)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test/endpoint", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	logOutput := logBuffer.String()

	// Should log when verbose is true
	if logBuffer.Len() == 0 {
		t.Error("expected logs when verbose=true, got nothing")
	}

	// Check log contains expected information
	expectedParts := []string{
		"[HTTP]",
		"GET",
		"/test/endpoint",
		"200",
		"13 bytes", // "test response" is 13 bytes
	}

	for _, part := range expectedParts {
		if !strings.Contains(logOutput, part) {
			t.Errorf("expected log to contain '%s', got: %s", part, logOutput)
		}
	}
}

func TestLoggingMiddleware_CapturesStatusCode(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(log.Writer())

	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"201 Created", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			handler := LoggingMiddleware(true)(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("test"))
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			logOutput := logBuffer.String()

			// Check that the log contains the numeric status code
			statusCodeStr := fmt.Sprintf(" - %d - ", tt.statusCode)
			if !strings.Contains(logOutput, statusCodeStr) {
				t.Errorf("expected log to contain status code %d, got: %s", tt.statusCode, logOutput)
			}
		})
	}
}

func TestLoggingMiddleware_CapturesMethod(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(log.Writer())

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			logBuffer.Reset()

			handler := LoggingMiddleware(true)(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			logOutput := logBuffer.String()
			if !strings.Contains(logOutput, method) {
				t.Errorf("expected log to contain method %s, got: %s", method, logOutput)
			}
		})
	}
}

func TestResponseWriter_WriteCapture(t *testing.T) {
	rw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     200,
	}

	data := []byte("hello world")
	n, err := rw.Write(data)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	if rw.written != int64(len(data)) {
		t.Errorf("expected written count %d, got %d", len(data), rw.written)
	}
}
