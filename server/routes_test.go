package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/auth"
)

// Remove the mockDB implementation from here since it's now in mocks_test.go

func TestRegisterRoutes(t *testing.T) {
	s := &Server{
		port: 8080,
		db:   &mockDB{healthStatus: true},
	}

	jwtManager := auth.NewJWTManager("test-secret")
	handler := s.RegisterRoutes(nil, *jwtManager, "local")

	if handler == nil {
		t.Error("RegisterRoutes() returned nil handler")
	}
}

func TestLiveCheck(t *testing.T) {
	s := &Server{
		port: 8080,
		db:   &mockDB{healthStatus: true},
	}

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	s.liveCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("liveCheck() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedFields := []string{"status", "version", "message"}
	for _, field := range expectedFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing field: %s", field)
		}
	}

	if response["status"] != "ok" {
		t.Errorf("liveCheck() status = %v, want %v", response["status"], "ok")
	}
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name         string
		healthStatus bool
		wantStatus   string
	}{
		{
			name:         "Healthy DB",
			healthStatus: true,
			wantStatus:   "up",
		},
		{
			name:         "Unhealthy DB",
			healthStatus: false,
			wantStatus:   "down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				port: 8080,
				db:   &mockDB{healthStatus: tt.healthStatus},
			}

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			s.healthHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("healthHandler() status = %v, want %v", w.Code, http.StatusOK)
			}

			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if status := response["status"]; status != tt.wantStatus {
				t.Errorf("healthHandler() status = %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		wantStatus int
	}{
		{
			name:       "Valid JSON",
			data:       map[string]string{"test": "data"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Empty JSON",
			data:       map[string]string{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := writeJSON(w, tt.wantStatus, tt.data)
			if err != nil {
				t.Errorf("writeJSON() error = %v", err)
			}

			if w.Code != tt.wantStatus {
				t.Errorf("writeJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("writeJSON() Content-Type = %v, want application/json", ct)
			}

			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}
		})
	}
}

func TestPublicRoutesRateLimiting(t *testing.T) {
	s := &Server{
		port: 8080,
		db:   &mockDB{healthStatus: true},
	}

	jwtManager := auth.NewJWTManager("test-secret")
	handler := s.RegisterRoutes(nil, *jwtManager, "local")
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test rate limiting on /live endpoint
	for i := 0; i < 55; i++ {
		resp, err := http.Get(server.URL + "/live")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if i < 50 {
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Request %d: expected %d, got %d", i, http.StatusOK, resp.StatusCode)
			}
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Request %d: expected %d, got %d", i, http.StatusTooManyRequests, resp.StatusCode)
			}
		}
	}
}

func TestAuthRoutesRateLimiting(t *testing.T) {
	s := &Server{
		port: 8080,
		db:   &mockDB{healthStatus: true},
	}

	jwtManager := auth.NewJWTManager("test-secret")
	handler := s.RegisterRoutes(nil, *jwtManager, "local")
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test rate limiting on /login endpoint
	for i := 0; i < 35; i++ {
		loginBody := strings.NewReader(`{"email": "test@test.com", "password": "test123"}`)
		resp, err := http.Post(server.URL+"/login", "application/json", loginBody)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if i < 30 {
			if resp.StatusCode == http.StatusTooManyRequests {
				t.Errorf("Request %d: unexpected rate limit", i)
			}
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Request %d: expected rate limit, got status %d", i, resp.StatusCode)
			}
		}
	}
}

func TestPrivateRoutesRateLimiting(t *testing.T) {
	s := &Server{
		port: 8080,
		db:   &mockDB{healthStatus: true},
	}

	jwtManager := auth.NewJWTManager("test-secret")
	token, _ := jwtManager.GenerateToken(uuid.New().String(), "test@email.com", uuid.New().String())
	handler := s.RegisterRoutes(nil, *jwtManager, "test")
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{}

	// Test rate limiting on /user/profile endpoint
	for i := 0; i < 20; i++ {
		req, err := http.NewRequest("GET", server.URL+"/user/profile", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if i < 15 {
			if resp.StatusCode == http.StatusTooManyRequests {
				t.Errorf("Request %d: unexpected rate limit", i)
			}
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Request %d: expected rate limit, got status %d", i, resp.StatusCode)
			}
		}
	}
}

func TestRateLimitReset(t *testing.T) {
	s := &Server{
		port: 8080,
		db:   &mockDB{healthStatus: true},
	}

	jwtManager := auth.NewJWTManager("test-secret")
	handler := s.RegisterRoutes(nil, *jwtManager, "local")
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make initial requests to trigger rate limit
	for i := 0; i < 25; i++ {
		http.Get(server.URL + "/live")
	}

	// Wait for rate limit to reset
	time.Sleep(1 * time.Second)

	// Try another request
	resp, err := http.Get(server.URL + "/live")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("Rate limit should have reset after waiting")
	}
}
