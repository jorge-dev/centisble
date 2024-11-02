package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jorge-dev/centsible/internal/auth"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

type contextKey string

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userIDKey := contextKey("user_id")
		emailKey := contextKey("email")

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, emailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
