package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"program-manager/utils"
)

// RequestLogger is a middleware that logs HTTP requests
type RequestLogger struct {
	next http.Handler
}

// NewRequestLogger creates a new request logger middleware
func NewRequestLogger(next http.Handler) *RequestLogger {
	return &RequestLogger{next: next}
}

// ServeHTTP implements the http.Handler interface
func (rl *RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create a response recorder to capture the status code
	rw := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	
	// Process the request
	rl.next.ServeHTTP(rw, r)
	
	duration := time.Since(start)
	
	// Log the request
	clientIP := rl.getClientIP(r)
	
	utils.LogRequest(
		r.Method,
		r.URL.Path,
		clientIP,
		rw.statusCode,
		duration,
		map[string]interface{}{
			"user_agent":    r.UserAgent(),
			"content_type":  r.Header.Get("Content-Type"),
			"content_length": r.ContentLength,
			"query_params":  r.URL.Query(),
		},
	)
	
	// Log errors for failed requests
	if rw.statusCode >= 400 {
		utils.LogError(
			"HTTP_REQUEST",
			fmt.Sprintf("HTTP %d error", rw.statusCode),
			"",
			fmt.Errorf("request failed with status %d", rw.statusCode),
			map[string]interface{}{
				"method":   r.Method,
				"path":     r.URL.Path,
				"client_ip": clientIP,
				"status":   rw.statusCode,
			},
		)
	}
}

// responseRecorder wraps http.ResponseWriter to capture status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// getClientIP extracts the client IP from the request
func (rl *RequestLogger) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fallback to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// RequestLoggingMiddleware returns a middleware function that logs requests
func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return NewRequestLogger(next)
}

// APIEndpointLogger is a specialized logger for API endpoints
type APIEndpointLogger struct {
	next http.Handler
}

// NewAPIEndpointLogger creates a new API endpoint logger
func NewAPIEndpointLogger(next http.Handler) *APIEndpointLogger {
	return &APIEndpointLogger{next: next}
}

// ServeHTTP implements the http.Handler interface for API endpoints
func (ael *APIEndpointLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Extract relevant information
	endpoint := r.URL.Path
	method := r.Method
	
	// Log API endpoint access
	utils.LogOperation(
		"API_ACCESS",
		fmt.Sprintf("API %s %s accessed", method, endpoint),
		"",
		map[string]interface{}{
			"endpoint":   endpoint,
			"method":     method,
			"user_agent": r.UserAgent(),
			"content_type": r.Header.Get("Content-Type"),
		},
	)

	// Create response recorder
	rw := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

	// Process request
	ael.next.ServeHTTP(rw, r)

	duration := time.Since(start)

	// Log API response
	utils.LogOperation(
		"API_RESPONSE",
		fmt.Sprintf("API %s %s completed", method, endpoint),
		"",
		map[string]interface{}{
			"endpoint": endpoint,
			"method":   method,
			"status":   rw.statusCode,
			"duration": duration.Milliseconds(),
		},
	)
}