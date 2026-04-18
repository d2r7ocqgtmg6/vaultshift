package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newDedupeLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newDedupeMockServer(secrets map[string]map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list") == "true" {
			keys := []string{}
			for k := range secrets {
				keys = append(keys, k)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": keys}})
			return
		}
		path := r.URL.Path
		for k, v := range secrets {
			if "/v1/"+k == path {
				json.NewEncoder(w).Encode(map[string]interface{}{"data": v})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestNewDeduper_MissingClient(t *testing.T) {
	l := newDedupeLogger(t)
	_, err := NewDeduper(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewDeduper_MissingLogger(t *testing.T) {
	svr := newDedupeMockServer(nil)
	defer svr.Close()
	c, _ := New(svr.URL, "token")
	_, err := NewDeduper(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestDedupe_FindsDuplicates(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/a": {"password": "abc123"},
		"secret/b": {"password": "abc123"},
		"secret/c": {"password": "unique"},
	}
	svr := newDedupeMockServer(secrets)
	defer svr.Close()
	c, _ := New(svr.URL, "token")
	l := newDedupeLogger(t)
	d, _ := NewDeduper(c, l, true)
	results, err := d.Dedupe("secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(results))
	}
	if len(results[0].Paths) != 2 {
		t.Errorf("expected 2 paths in duplicate group, got %d", len(results[0].Paths))
	}
}

func TestDedupe_NoDuplicates(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/x": {"key": "val1"},
		"secret/y": {"key": "val2"},
	}
	svr := newDedupeMockServer(secrets)
	defer svr.Close()
	c, _ := New(svr.URL, "token")
	l := newDedupeLogger(t)
	d, _ := NewDeduper(c, l, false)
	results, err := d.Dedupe("secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no duplicates, got %d", len(results))
	}
}
