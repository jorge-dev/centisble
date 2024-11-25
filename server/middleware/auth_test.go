package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jorge-dev/centsible/internal/auth"
)

func TestNewAuthMiddleware(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	if middleware == nil {
		t.Error("NewAuthMiddleware() returned nil")
	} else if middleware.jwtManager != jwtManager {
		t.Error("NewAuthMiddleware() did not set correct jwtManager")
	}
}

func TestAuthRequired(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		checkContext   bool
	}{
		{
			name: "Valid token",
			setupAuth: func() string {
				token, _ := jwtManager.GenerateToken("123", "test@example.com", "user")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name: "Missing Authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Invalid token format",
			setupAuth: func() string {
				return "Bearer invalid.token.format"
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Token without Bearer prefix",
			setupAuth: func() string {
				token, _ := jwtManager.GenerateToken("123", "test@example.com", "user")
				return token
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var contextUserID, contextEmail, contextRoleID string

			// Create test handler that checks context values
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					contextUserID = r.Context().Value(UserIDKey).(string)
					contextEmail = r.Context().Value(EmailKey).(string)
					contextRoleID = r.Context().Value(RoleIDKey).(string)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set("Authorization", auth)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create and execute middleware
			handler := middleware.AuthRequired(nextHandler)
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}

			// Verify context values for successful cases
			if tt.checkContext && tt.expectedStatus == http.StatusOK {
				if contextUserID != "123" {
					t.Errorf("context user ID = %v, want %v", contextUserID, "123")
				}
				if contextEmail != "test@example.com" {
					t.Errorf("context email = %v, want %v", contextEmail, "test@example.com")
				}
				if contextRoleID != "user" {
					t.Errorf("context role ID = %v, want %v", contextRoleID, "user")
				}
			}
		})
	}
}

func TestContextKeys(t *testing.T) {
	tests := []struct {
		key  contextKey
		want string
	}{
		{UserIDKey, "user_id"},
		{EmailKey, "email"},
		{RoleIDKey, "role_id"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			if string(tt.key) != tt.want {
				t.Errorf("contextKey = %v, want %v", tt.key, tt.want)
			}
		})
	}
}
