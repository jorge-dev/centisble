package auth

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	InactivityTimeout    = 30 * time.Minute
	claimIssuer          = "centisble-auth"
	claimAudience        = "centisble-api"
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type JWTClaims struct {
	UserID    string `json:"user_id"`
	RoleIDKey string `json:"role_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey []byte
	blacklist map[string]time.Time // Simple in-memory blacklist
	sessions  map[string]time.Time // Track last activity for each user
	timeout   time.Duration        // Add this field
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		blacklist: make(map[string]time.Time),
		sessions:  make(map[string]time.Time),
		timeout:   InactivityTimeout, // Set default timeout
	}
}

// Add this method for testing
func (m *JWTManager) SetTimeout(duration time.Duration) {
	m.timeout = duration
}

func (m *JWTManager) GenerateTokenPair(userID, email, roleID string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := m.generateAccessToken(userID, email, roleID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token with longer expiration
	refreshToken, err := m.generateRefreshToken(userID, email, roleID)
	if err != nil {
		slog.Debug("Error generating refresh token", "error", err)
		return nil, err
	}

	expiresAt := time.Now().Add(AccessTokenDuration)

	// Update session activity
	m.sessions[userID] = time.Now()

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (m *JWTManager) generateAccessToken(userID, email, roleID string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		RoleIDKey: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    claimIssuer,
			Audience:  []string{claimAudience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) generateRefreshToken(userID, email, roleID string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		RoleIDKey: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    claimIssuer,
			Audience:  []string{"centisble-api-refresh"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify this is a refresh token
	if !contains(claims.Audience, "centisble-api-refresh") {
		return nil, fmt.Errorf("invalid token type for refresh")
	}

	// Check for session timeout
	lastActivity, exists := m.sessions[claims.UserID]
	if !exists || time.Since(lastActivity) > InactivityTimeout {
		delete(m.sessions, claims.UserID) // Clear the session
		return nil, fmt.Errorf("session expired due to inactivity")
	}

	// Generate new token pair
	return m.GenerateTokenPair(claims.UserID, claims.Email, claims.RoleIDKey)
}

func (m *JWTManager) UpdateActivity(userID string) {
	m.sessions[userID] = time.Now()
}

func (m *JWTManager) GenerateToken(userID, email, roleID string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		RoleIDKey: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    claimIssuer,
			Audience:  []string{claimAudience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) InvalidateToken(tokenString string) error {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	// Add to blacklist until expiration
	m.blacklist[tokenString] = claims.ExpiresAt.Time
	return nil
}

func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Check if token is blacklisted
	if _, blacklisted := m.blacklist[tokenString]; blacklisted {
		return nil, fmt.Errorf("token has been invalidated")
	}

	// Clean up expired blacklisted tokens and inactive sessions
	m.cleanupBlacklist()
	m.cleanupSessions()

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation
	if claims.Issuer != claimIssuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	if !contains(claims.Audience, claimAudience) {
		return nil, fmt.Errorf("invalid token audience")
	}

	// Check for session timeout if it's an access token
	if contains(claims.Audience, claimAudience) {
		lastActivity, exists := m.sessions[claims.UserID]
		if !exists || time.Since(lastActivity) > m.timeout { // Use m.timeout instead of InactivityTimeout
			delete(m.sessions, claims.UserID) // Clear the session
			return nil, fmt.Errorf("session expired due to inactivity")
		}
		// Update last activity
		m.UpdateActivity(claims.UserID)
	}

	return claims, nil
}

func (m *JWTManager) cleanupBlacklist() {
	now := time.Now()
	for token, expiry := range m.blacklist {
		if now.After(expiry) {
			delete(m.blacklist, token)
		}
	}
}

func (m *JWTManager) cleanupSessions() {
	now := time.Now()
	for userID, lastActivity := range m.sessions {
		if now.Sub(lastActivity) > InactivityTimeout {
			delete(m.sessions, userID)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
