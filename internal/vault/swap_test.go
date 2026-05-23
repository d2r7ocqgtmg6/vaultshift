package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newSwapLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newSwapMockServer(t *testing.T, store map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if inner, ok := body["data"].(map[string]interface{}); ok {
				store[path] = inner
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func TestNewSwapper_MissingClient(t *testing.T) {
	l := newSwapLogger(t)
	_, err := NewSwapper(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewSwapper_MissingLogger(t *testing.T) {
	c := &Client{}
	_, err := NewSwapper(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestSwap_DryRun_NoWrite(t *testing.T) {
	store := map[string]map[string]interface{}{
		"/v1/secret/data/a": {"key": "alpha"},
		"/v1/secret/data/b": {"key": "beta"},
	}
	srv := newSwapMockServer(t, store)
	defer srv.Close()

	c, err := New(srv.URL, "token", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l := newSwapLogger(t)
	sw, err := NewSwapper(c, l, true)
	if err != nil {
		t.Fatalf("NewSwapper: %v", err)
	}
	if err := sw.Swap("secret/data/a", "secret/data/b"); err != nil {
		t.Fatalf("Swap: %v", err)
	}
	if v := store["/v1/secret/data/a"]["key"]; v != "alpha" {
		t.Errorf("dry-run modified path a: got %v", v)
	}
}

func TestSwap_ExchangesValues(t *testing.T) {
	store := map[string]map[string]interface{}{
		"/v1/secret/data/a": {"key": "alpha"},
		"/v1/secret/data/b": {"key": "beta"},
	}
	srv := newSwapMockServer(t, store)
	defer srv.Close()

	c, err := New(srv.URL, "token", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l := newSwapLogger(t)
	sw, err := NewSwapper(c, l, false)
	if err != nil {
		t.Fatalf("NewSwapper: %v", err)
	}
	if err := sw.Swap("secret/data/a", "secret/data/b"); err != nil {
		t.Fatalf("Swap: %v", err)
	}
	if v := store["/v1/secret/data/a"]["key"]; v != "beta" {
		t.Errorf("path a: want beta, got %v", v)
	}
	if v := store["/v1/secret/data/b"]["key"]; v != "alpha" {
		t.Errorf("path b: want alpha, got %v", v)
	}
}
