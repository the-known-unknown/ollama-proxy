package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const (
	headerAuthorization = "Authorization"
	headerAPIKey        = "X-API-Key"
	bearerPrefix        = "Bearer "
)

func Middleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if expectedKey == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !validKey(r, expectedKey) {
				unauthorized(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func validKey(r *http.Request, expectedKey string) bool {
	provided := extractKey(r)
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expectedKey)) == 1
}

func extractKey(r *http.Request) string {
	if h := r.Header.Get(headerAuthorization); strings.HasPrefix(h, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))
	}
	return strings.TrimSpace(r.Header.Get(headerAPIKey))
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}
