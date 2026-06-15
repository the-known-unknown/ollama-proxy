package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestPingFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	if err := client.Ping(context.Background()); err == nil {
		t.Fatal("expected ping error")
	}
}

func TestListModelsAndHasModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"llama3"},{"id":"mistral"}]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("got %d models, want 2", len(models))
	}

	ok, err := client.HasModel(context.Background(), "mistral")
	if err != nil || !ok {
		t.Errorf("expected mistral to exist (ok=%v err=%v)", ok, err)
	}
	ok, _ = client.HasModel(context.Background(), "ghost")
	if ok {
		t.Error("did not expect ghost to exist")
	}
}
