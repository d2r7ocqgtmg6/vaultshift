package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newRenameLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newRenameMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]any{
		"secret/data/old": {"key": "value"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			store[path] = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			delete(store, path)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestRename_DryRun_NoWrite(t *testing.T) {
	srv := newRenameMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newRenameLogger(t)
	r, _ := NewRenamer(c, l, true)
	if err := r.Rename("secret/data/old", "secret/data/new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRename_MovesSecret(t *testing.T) {
	srv := newRenameMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newRenameLogger(t)
	r, _ := NewRenamer(c, l, false)
	if err := r.Rename("secret/data/old", "secret/data/new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRenamer_MissingClient(t *testing.T) {
	l := newRenameLogger(t)
	_, err := NewRenamer(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}
