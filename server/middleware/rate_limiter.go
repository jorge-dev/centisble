package middleware

import (
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters sync.Map // map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		rate:  r,
		burst: b,
	}
}

func (rl *RateLimiter) getClientLimiter(clientIP string) *rate.Limiter {
	if limiter, exists := rl.limiters.Load(clientIP); exists {
		return limiter.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters.Store(clientIP, limiter)
	return limiter
}

func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fall back to RemoteAddr if X-Forwarded-For is not present
	ip := r.RemoteAddr
	// Remove port number if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := rl.getClientIP(r)
		limiter := rl.getClientLimiter(clientIP)

		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
