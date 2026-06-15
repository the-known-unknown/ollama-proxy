package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPlatformRunning(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := NewOllamaPlatform(NewClient(srv.URL))
	if !p.Running(context.Background()) {
		t.Fatal("expected platform to be running")
	}
}

func TestEnsureRunningWhenAlreadyUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := NewOllamaPlatform(NewClient(srv.URL))
	started, err := p.EnsureRunning(context.Background(), time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("ensure running: %v", err)
	}
	if started {
		t.Error("did not expect to start an already-running platform")
	}
}

func TestWaitReadyTimeout(t *testing.T) {
	p := NewOllamaPlatform(NewClient("http://127.0.0.1:0"))
	if err := p.WaitReady(context.Background(), 200*time.Millisecond, 50*time.Millisecond); err == nil {
		t.Fatal("expected timeout error")
	}
}
