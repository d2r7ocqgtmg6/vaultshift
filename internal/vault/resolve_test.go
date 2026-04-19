package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func newResolveMockServer(t *testing.T, path string, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/"+path {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
			return
		}
		http.NotFound(w, r)
	}))
}

func newResolveClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c
}

func TestNewResolver_MissingClient(t *testing.T) {
	_, err := NewResolver(nil, zerolog.Nop(), false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestResolve_KeyFound(t *testing.T) {
	srv := newResolveMockServer(t, "secret/data/app", map[string]interface{}{"api_key": "abc123"})
	defer srv.Close()
	client := newResolveClient(t, srv)

	r, err := NewResolver(client, zerolog.Nop(), false)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	val, err := r.Resolve("secret/data/app", "api_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "abc123" {
		t.Errorf("expected abc123, got %s", val)
	}
}

func TestResolve_KeyMissing(t *testing.T) {
	srv := newResolveMockServer(t, "secret/data/app", map[string]interface{}{"other": "val"})
	defer srv.Close()
	client := newResolveClient(t, srv)

	r, _ := NewResolver(client, zerolog.Nop(), false)
	_, err := r.Resolve("secret/data/app", "missing_key")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestResolveMap_ReturnsAllKeys(t *testing.T) {
	srv := newResolveMockServer(t, "secret/data/cfg", map[string]interface{}{"a": "1", "b": "2"})
	defer srv.Close()
	client := newResolveClient(t, srv)

	r, _ := NewResolver(client, zerolog.Nop(), false)
	m, err := r.ResolveMap("secret/data/cfg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["a"] != "1" || m["b"] != "2" {
		t.Errorf("unexpected map: %v", m)
	}
}
