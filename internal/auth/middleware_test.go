package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func nextHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestMiddlewareDisabledWhenNoKey(t *testing.T) {
	h := Middleware("")(nextHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestMiddlewareAuth(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		value      string
		wantStatus int
	}{
		{"valid bearer", "Authorization", "Bearer secret", http.StatusOK},
		{"valid x-api-key", "X-API-Key", "secret", http.StatusOK},
		{"wrong bearer", "Authorization", "Bearer nope", http.StatusUnauthorized},
		{"wrong x-api-key", "X-API-Key", "nope", http.StatusUnauthorized},
		{"missing", "", "", http.StatusUnauthorized},
		{"empty bearer", "Authorization", "Bearer ", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Middleware("secret")(nextHandler())
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(tt.header, tt.value)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
