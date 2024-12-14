package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name          string
		rateLimit     rate.Limit
		burst         int
		requestCount  int
		expectedCodes []int
		delay         time.Duration
	}{
		{
			name:          "Allow requests within rate limit",
			rateLimit:     2,
			burst:         2,
			requestCount:  2,
			expectedCodes: []int{200, 200},
			delay:         0,
		},
		{
			name:          "Block requests exceeding rate limit",
			rateLimit:     1,
			burst:         1,
			requestCount:  3,
			expectedCodes: []int{200, 429, 429},
			delay:         0,
		},
		{
			name:          "Allow burst then block",
			rateLimit:     1,
			burst:         2,
			requestCount:  4,
			expectedCodes: []int{200, 200, 429, 429},
			delay:         0,
		},
		{
			name:          "Allow requests with delay",
			rateLimit:     2,
			burst:         1,
			requestCount:  3,
			expectedCodes: []int{200, 200, 200},
			delay:         time.Second / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that always returns 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create the rate limiter middleware
			limiter := NewRateLimiter(tt.rateLimit, tt.burst)
			handler := limiter.Limit(nextHandler)

			// Make requests and check responses
			for i := 0; i < tt.requestCount; i++ {
				// Create a test request
				req := httptest.NewRequest("GET", "/test", nil)
				rr := httptest.NewRecorder()

				// Handle the request
				handler.ServeHTTP(rr, req)

				// Check if the status code matches expected
				if rr.Code != tt.expectedCodes[i] {
					t.Errorf("Request %d: expected status code %d, got %d",
						i+1, tt.expectedCodes[i], rr.Code)
				}

				// Wait if delay is specified
				if tt.delay > 0 {
					time.Sleep(tt.delay)
				}
			}
		})
	}
}

func TestRateLimiterClientIdentification(t *testing.T) {
	limiter := NewRateLimiter(1, 1)
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name          string
		headers       map[string]string
		remoteAddr    string
		requestCount  int
		expectedCodes []int
	}{
		{
			name: "Different IPs should have separate limits",
			headers: map[string]string{
				"X-Forwarded-For": "1.2.3.4",
			},
			requestCount:  2,
			expectedCodes: []int{200, 429},
		},
		{
			name: "Multiple requests from same IP",
			headers: map[string]string{
				"X-Forwarded-For": "5.6.7.8",
			},
			requestCount:  2,
			expectedCodes: []int{200, 429},
		},
		{
			name:          "No X-Forwarded-For header",
			headers:       map[string]string{},
			remoteAddr:    "9.10.11.12:1234",
			requestCount:  2,
			expectedCodes: []int{200, 429},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.requestCount; i++ {
				req := httptest.NewRequest("GET", "/test", nil)

				// Set headers
				for key, value := range tt.headers {
					req.Header.Set(key, value)
				}

				// Set RemoteAddr if specified
				if tt.remoteAddr != "" {
					req.RemoteAddr = tt.remoteAddr
				}

				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				if rr.Code != tt.expectedCodes[i] {
					t.Errorf("Request %d: expected status code %d, got %d",
						i+1, tt.expectedCodes[i], rr.Code)
				}
			}
		})
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	limiter := NewRateLimiter(5, 5)
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a channel to collect results
	results := make(chan int, 10)

	// Make 10 concurrent requests
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			results <- rr.Code
		}()
	}

	// Collect results
	successCount := 0
	rateLimitCount := 0

	for i := 0; i < 10; i++ {
		code := <-results
		switch code {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitCount++
		}
	}

	// Verify that we got the expected mix of responses
	if successCount != 5 {
		t.Errorf("expected 5 successful requests, got %d", successCount)
	}
	if rateLimitCount != 5 {
		t.Errorf("expected 5 rate limited requests, got %d", rateLimitCount)
	}
}
