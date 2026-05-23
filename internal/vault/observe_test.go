package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newObserveMockServer(t *testing.T, paths map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for k, v := range paths {
			if path == "/v1/"+k {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"data": v})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newObserveClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return c
}

func newObserveLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	return l
}

func TestNewObserver_MissingClient(t *testing.T) {
	l := newObserveLogger(t)
	_, err := NewObserver(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewObserver_MissingLogger(t *testing.T) {
	srv := newObserveMockServer(t, nil)
	defer srv.Close()
	c := newObserveClient(t, srv)
	_, err := NewObserver(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestObserve_ExistingPath(t *testing.T) {
	srv := newObserveMockServer(t, map[string]map[string]interface{}{
		"secret/data/app": {"key": "value"},
	})
	defer srv.Close()
	c := newObserveClient(t, srv)
	l := newObserveLogger(t)
	obs, err := NewObserver(c, l, false)
	if err != nil {
		t.Fatalf("NewObserver: %v", err)
	}
	results := obs.Observe(context.Background(), []string{"secret/data/app"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Exists {
		t.Error("expected Exists=true")
	}
	if results[0].KeyCount != 1 {
		t.Errorf("expected KeyCount=1, got %d", results[0].KeyCount)
	}
}

func TestObserve_MissingPath(t *testing.T) {
	srv := newObserveMockServer(t, map[string]map[string]interface{}{})
	defer srv.Close()
	c := newObserveClient(t, srv)
	l := newObserveLogger(t)
	obs, _ := NewObserver(c, l, false)
	results := obs.Observe(context.Background(), []string{"secret/data/missing"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Exists {
		t.Error("expected Exists=false for missing path")
	}
}
