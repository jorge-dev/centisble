package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name           string
		existingHeader string
		existingValue  string
		expectedHeader string
		expectedValue  string
		shouldOverride bool
	}{
		{
			name:           "Sets X-Content-Type-Options",
			expectedHeader: "X-Content-Type-Options",
			expectedValue:  "nosniff",
			shouldOverride: false,
		},
		{
			name:           "Sets X-Frame-Options",
			expectedHeader: "X-Frame-Options",
			expectedValue:  "DENY",
			shouldOverride: false,
		},
		{
			name:           "Sets X-XSS-Protection",
			expectedHeader: "X-XSS-Protection",
			expectedValue:  "1; mode=block",
			shouldOverride: false,
		},
		{
			name:           "Sets Strict-Transport-Security",
			expectedHeader: "Strict-Transport-Security",
			expectedValue:  "max-age=31536000; includeSubDomains",
			shouldOverride: false,
		},
		{
			name:           "Sets Content-Security-Policy",
			expectedHeader: "Content-Security-Policy",
			expectedValue:  "default-src 'self'",
			shouldOverride: false,
		},
		{
			name:           "Sets Referrer-Policy",
			expectedHeader: "Referrer-Policy",
			expectedValue:  "strict-origin-when-cross-origin",
			shouldOverride: false,
		},
		{
			name:           "Sets Permissions-Policy",
			expectedHeader: "Permissions-Policy",
			expectedValue:  "geolocation=(), camera=(), microphone=()",
			shouldOverride: false,
		},
		{
			name:           "Respects existing X-Content-Type-Options",
			existingHeader: "X-Content-Type-Options",
			existingValue:  "custom-value",
			expectedHeader: "X-Content-Type-Options",
			expectedValue:  "custom-value",
			shouldOverride: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new security headers middleware
			security := NewSecurityHeaders()

			// Create a test handler that we'll wrap with our middleware
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handler does nothing
			})

			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// If we're testing with existing headers, set them
			if tt.existingHeader != "" {
				rr.Header().Set(tt.existingHeader, tt.existingValue)
			}

			// Run the middleware
			handler := security.Handler(nextHandler)
			handler.ServeHTTP(rr, req)

			// Check if the header was set correctly
			gotValue := rr.Header().Get(tt.expectedHeader)
			if tt.shouldOverride {
				if gotValue != tt.expectedValue {
					t.Errorf("handler didn't override header %s: got %v want %v",
						tt.expectedHeader, gotValue, tt.expectedValue)
				}
			} else {
				expectedValue := tt.expectedValue
				if tt.existingHeader != "" {
					expectedValue = tt.existingValue
				}
				if gotValue != expectedValue {
					t.Errorf("handler set wrong value for %s: got %v want %v",
						tt.expectedHeader, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestAllSecurityHeadersPresent(t *testing.T) {
	security := NewSecurityHeaders()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := security.Handler(nextHandler)
	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "geolocation=(), camera=(), microphone=()",
	}

	for header, expectedValue := range expectedHeaders {
		if got := rr.Header().Get(header); got != expectedValue {
			t.Errorf("missing or incorrect security header %s: got %v want %v",
				header, got, expectedValue)
		}
	}
}
