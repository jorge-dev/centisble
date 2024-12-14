package middleware

import "net/http"

type SecurityHeaders struct{}

func NewSecurityHeaders() *SecurityHeaders {
	return &SecurityHeaders{}
}

func (s *SecurityHeaders) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set security headers if they haven't been set yet
		headers := w.Header()
		if headers.Get("X-Content-Type-Options") == "" {
			headers.Set("X-Content-Type-Options", "nosniff")
		}
		if headers.Get("X-Frame-Options") == "" {
			headers.Set("X-Frame-Options", "DENY")
		}
		if headers.Get("X-XSS-Protection") == "" {
			headers.Set("X-XSS-Protection", "1; mode=block")
		}
		if headers.Get("Strict-Transport-Security") == "" {
			headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if headers.Get("Content-Security-Policy") == "" {
			headers.Set("Content-Security-Policy", "default-src 'self'")
		}
		if headers.Get("Referrer-Policy") == "" {
			headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		}
		if headers.Get("Permissions-Policy") == "" {
			headers.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}
