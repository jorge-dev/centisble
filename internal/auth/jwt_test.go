package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

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
			email:   "test@example.com",
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
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
		check     func(*JWTClaims) error
	}{
		{
			name: "Valid token",
			setupFunc: func() string {
				token, _ := manager.GenerateToken("123", "test@example.com", "user")
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
					Email:     "test@example.com",
					RoleIDKey: "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, _ := token.SignedString([]byte("test-secret"))
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
					Email:     "test@example.com",
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
	secretKey := "test-secret"
	manager := NewJWTManager(secretKey)

	if manager == nil {
		t.Error("NewJWTManager() returned nil")
	}

	if string(manager.secretKey) != secretKey {
		t.Errorf("NewJWTManager() secretKey = %v, want %v", string(manager.secretKey), secretKey)
	}
}
