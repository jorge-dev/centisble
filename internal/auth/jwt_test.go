package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const secretKey = "test-secret"
const testEmail = "test@example.com"

func TestJWTManagerGenerateToken(t *testing.T) {
	manager := NewJWTManager(secretKey)
	// Add cleanup
	defer func() {
		manager.sessions = make(map[string]time.Time)
		manager.blacklist = make(map[string]time.Time)
	}()

	tests := []struct {
		name    string
		userID  string
		email   string
		roleID  string
		wantErr bool
	}{
		{
			name:    "Valid token generation",
			userID:  "123",
			email:   testEmail,
			roleID:  "user",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tt.userID, tt.email, tt.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
				// Initialize session after token generation
				manager.UpdateActivity(tt.userID)
			}
		})
	}
}

func TestJWTManagerValidateToken(t *testing.T) {
	manager := NewJWTManager(secretKey)
	// Add cleanup after tests
	defer func() {
		manager.sessions = make(map[string]time.Time)
		manager.blacklist = make(map[string]time.Time)
	}()

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
		check     func(*JWTClaims) error
	}{
		{
			name: "Valid token",
			setupFunc: func() string {
				// Generate token
				token, _ := manager.GenerateToken("123", testEmail, "user")
				// Initialize session for the user
				manager.UpdateActivity("123")
				return token
			},
			wantErr: false,
			check: func(claims *JWTClaims) error {
				if claims.UserID != "123" {
					return fmt.Errorf("invalid user ID in claims")
				}
				return nil
			},
		},
		{
			name: "Expired token",
			setupFunc: func() string {
				claims := &JWTClaims{
					UserID:    "123",
					Email:     testEmail,
					RoleIDKey: "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, _ := token.SignedString([]byte(secretKey))
				return signedToken
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "Invalid signing method",
			setupFunc: func() string {
				claims := &JWTClaims{
					UserID:    "123",
					Email:     testEmail,
					RoleIDKey: "user",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
				signedToken, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				return signedToken
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "Invalid token format",
			setupFunc: func() string {
				return "invalid.token.format"
			},
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupFunc()
			claims, err := manager.ValidateToken(token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				if err := tt.check(claims); err != nil {
					t.Errorf("Claim validation failed: %v", err)
				}
			}
		})
	}
}

func TestNewJWTManager(t *testing.T) {
	manager := NewJWTManager(secretKey)

	if manager == nil {
		t.Error("NewJWTManager() returned nil")
	}

	if manager != nil && string(manager.secretKey) != secretKey {
		t.Errorf("NewJWTManager() secretKey = %v, want %v", string(manager.secretKey), secretKey)
	}
}

func TestJWTManager(t *testing.T) {
	manager := NewJWTManager(secretKey)
	// Add cleanup
	defer func() {
		manager.sessions = make(map[string]time.Time)
		manager.blacklist = make(map[string]time.Time)
	}()

	t.Run("GenerateToken", func(t *testing.T) {
		// Test data
		userID := "123"
		email := testEmail
		roleID := "user"

		token, err := manager.GenerateToken(userID, email, roleID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Initialize session after generating token
		manager.UpdateActivity(userID)

		// Validate the generated token
		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roleID, claims.RoleIDKey)
		assert.Equal(t, claimIssuer, claims.Issuer)
		assert.Contains(t, claims.Audience, claimAudience)
	})

	t.Run("ValidateToken_InvalidToken", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("ValidateToken_ExpiredToken", func(t *testing.T) {
		// Create a token that's already expired
		claims := JWTClaims{
			UserID:    "123",
			Email:     testEmail,
			RoleIDKey: "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    claimIssuer,
				Audience:  []string{claimAudience},
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		assert.NoError(t, err)

		_, err = manager.ValidateToken(tokenString)
		assert.Error(t, err)
	})

	t.Run("InvalidateToken", func(t *testing.T) {
		// Generate a valid token
		token, err := manager.GenerateToken("123", testEmail, "user")
		assert.NoError(t, err)

		// Validate it works before invalidation
		_, err = manager.ValidateToken(token)
		assert.NoError(t, err)

		// Invalidate the token
		err = manager.InvalidateToken(token)
		assert.NoError(t, err)

		// Try to validate the invalidated token
		_, err = manager.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has been invalidated")
	})

	t.Run("BlacklistCleanup", func(t *testing.T) {
		manager := NewJWTManager(secretKey)

		// Create an expired token
		expiredClaims := &JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		}

		// Add expired token to blacklist
		expiredToken := "expired-token"
		manager.blacklist[expiredToken] = expiredClaims.ExpiresAt.Time

		// Trigger cleanup by validating any token
		manager.cleanupBlacklist()

		// Check if expired token was removed
		_, exists := manager.blacklist[expiredToken]
		assert.False(t, exists, "Expired token should be removed from blacklist")
	})

	t.Run("ValidateToken_InvalidIssuer", func(t *testing.T) {
		// Create a token with invalid issuer
		claims := JWTClaims{
			UserID:    "123",
			Email:     testEmail,
			RoleIDKey: "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "wrong-issuer",
				Audience:  []string{claimAudience},
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		assert.NoError(t, err)

		_, err = manager.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token issuer")
	})

	t.Run("ValidateToken_InvalidAudience", func(t *testing.T) {
		// Create a token with invalid audience
		claims := JWTClaims{
			UserID:    "123",
			Email:     testEmail,
			RoleIDKey: "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    claimIssuer,
				Audience:  []string{"wrong-audience"},
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		assert.NoError(t, err)

		_, err = manager.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token audience")
	})
}
