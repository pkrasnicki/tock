package http

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggingMiddleware logs HTTP requests with configurable verbosity
func LoggingMiddleware(verbose bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !verbose {
				next(w, r)
				return
			}

			start := time.Now()
			
			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status
			}

			// Call the next handler
			next(wrapped, r)

			// Log after the request is complete
			duration := time.Since(start)
			
			log.Printf("[HTTP] %s %s - %d - %v - %s - %d bytes",
				r.Method,
				r.RequestURI,
				wrapped.statusCode,
				duration,
				r.RemoteAddr,
				wrapped.written,
			)
		}
	}
}
