package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestProxyForwardsAndStripsAuth(t *testing.T) {
	var gotAuth, gotAPIKey, gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAPIKey = r.Header.Get("X-API-Key")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello from upstream"))
	}))
	defer upstream.Close()

	target, _ := url.Parse(upstream.URL)
	front := httptest.NewServer(New(target))
	defer front.Close()

	req, _ := http.NewRequest(http.MethodGet, front.URL+"/v1/models", nil)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-API-Key", "secret")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "hello from upstream" {
		t.Errorf("body = %q", body)
	}
	if gotPath != "/v1/models" {
		t.Errorf("path = %q, want /v1/models", gotPath)
	}
	if gotAuth != "" {
		t.Errorf("expected Authorization header to be stripped, got %q", gotAuth)
	}
	if gotAPIKey != "" {
		t.Errorf("expected X-API-Key header to be stripped, got %q", gotAPIKey)
	}
}
