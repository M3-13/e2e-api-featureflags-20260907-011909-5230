package main

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// RequireAuth guards every route except GET /healthz with an API key check.
// The key is read from the X-API-Key header and compared against the API_KEY
// environment variable in constant time via crypto/subtle. When the key is
// missing, empty, or wrong, it answers 401 with {"error":"unauthorized"}.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		expected := os.Getenv("API_KEY")
		provided := r.Header.Get("X-API-Key")

		// Without a configured key, refuse everything rather than accept an
		// empty key by accident.
		if expected == "" {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
