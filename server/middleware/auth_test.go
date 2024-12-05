package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		expectedError  string
		checkContext   bool
	}{
		{
			name: "Valid token",
			setupAuth: func() string {
				tokenPair, _ := jwtManager.GenerateTokenPair("123", "test@example.com", "user")
				return "Bearer " + tokenPair.AccessToken
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
			expectedError:  "Authorization header required",
			checkContext:   false,
		},

		{
			name: "Invalid token format",
			setupAuth: func() string {
				return "Bearer invalid.token.format"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
			checkContext:   false,
		},

		{
			name: "Expired token",
			setupAuth: func() string {
				claims := auth.JWTClaims{
					UserID:    "123",
					Email:     "test@example.com",
					RoleIDKey: "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						Issuer:    "centisble-auth",
						Audience:  []string{"centisble-api"},
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Access token has expired, please refresh",
			checkContext:   false,
		},

		{
			name: "Token without Bearer prefix",
			setupAuth: func() string {
				tokenPair, _ := jwtManager.GenerateTokenPair("123", "test@example.com", "user")
				return tokenPair.AccessToken
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},

		{
			name: "Inactive session",
			setupAuth: func() string {
				tokenPair, _ := jwtManager.GenerateTokenPair("123", "test@example.com", "user")
				jwtManager.SetTimeout(10 * time.Millisecond) // Set very short timeout for testing

				// Force session expiration by manipulating the session time
				jwtManager.UpdateActivity("123")
				time.Sleep(20 * time.Millisecond) // Wait just longer than timeout
				return "Bearer " + tokenPair.AccessToken
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Session expired, please login again",
			checkContext:   false,
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

			if tt.expectedError != "" {
				var response map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}
				if message, exists := response["message"]; !exists || !contains(message, tt.expectedError) {
					t.Errorf("Expected error message containing %q, got %q", tt.expectedError, message)
				}
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

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
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
