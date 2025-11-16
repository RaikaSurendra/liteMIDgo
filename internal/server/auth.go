package server

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuth middleware for protecting endpoints
func (s *Server) BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health endpoint
		if r.URL.Path == "/health" {
			next(w, r)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="LiteMIDgo"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		validUsername := s.config.Server.Auth.Username
		validPassword := s.config.Server.Auth.Password

		if subtle.ConstantTimeCompare([]byte(username), []byte(validUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(password), []byte(validPassword)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="LiteMIDgo"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// SecurityHeaders middleware for adding security headers
func (s *Server) SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next(w, r)
	}
}
