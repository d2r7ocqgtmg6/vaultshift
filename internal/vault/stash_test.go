package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newStashMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func newStashLogger(t *testing.T) AuditLogger {
	t.Helper()
	l, err := New("")
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	return l
}

func TestNewStasher_MissingClient(t *testing.T) {
	_, err := NewStasher(nil, newStashLogger(t), false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewStasher_MissingLogger(t *testing.T) {
	srv := newStashMockServer(t)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "tok"})
	_, err := NewStasher(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestStash_DryRun_NoFile(t *testing.T) {
	srv := newStashMockServer(t)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "tok"})
	s, _ := NewStasher(c, newStashLogger(t), true)

	entries, err := s.Stash([]string{"secret/a"}, "/tmp/should-not-exist.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if _, err := os.Stat("/tmp/should-not-exist.jsonl"); !os.IsNotExist(err) {
		os.Remove("/tmp/should-not-exist.jsonl")
		t.Fatal("file should not be created in dry-run mode")
	}
}

func TestStash_SaveAndLoad(t *testing.T) {
	srv := newStashMockServer(t)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "tok"})
	s, _ := NewStasher(c, newStashLogger(t), false)

	out := filepath.Join(t.TempDir(), "stash.jsonl")
	entries, err := s.Stash([]string{"secret/a", "secret/b"}, out)
	if err != nil {
		t.Fatalf("stash: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	loaded, err := LoadStash(out)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 loaded entries, got %d", len(loaded))
	}
	if loaded[0].Path != "secret/a" {
		t.Errorf("unexpected path: %s", loaded[0].Path)
	}
}

func TestLoadStash_InvalidFile(t *testing.T) {
	_, err := LoadStash("/nonexistent/stash.jsonl")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
