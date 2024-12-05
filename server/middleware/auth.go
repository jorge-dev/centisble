package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jorge-dev/centsible/internal/auth"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
	RoleIDKey contextKey = "role_id"
)

type contextKey string

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "Authorization header required",
			})
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)

			switch {
			case strings.Contains(err.Error(), "inactivity"):
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "session_expired",
					"message": "Session expired, please login again",
				})
			case strings.Contains(err.Error(), "expired"):
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "token_expired",
					"message": "Access token has expired, please refresh",
				})
			default:
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "invalid_token",
					"message": "Invalid token",
				})
			}
			return
		}

		// Update user's last activity
		m.jwtManager.UpdateActivity(claims.UserID)

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, RoleIDKey, claims.RoleIDKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
