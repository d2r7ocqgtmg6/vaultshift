package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newArchiveLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newArchiveMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "secret/data/prod/db"):
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"data": map[string]any{"password": "s3cr3t"}},
			})
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestNewArchiver_MissingClient(t *testing.T) {
	l := newArchiveLogger(t)
	_, err := NewArchiver(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestArchive_DryRun_NoDelete(t *testing.T) {
	srv := newArchiveMockServer(t)
	defer srv.Close()

	c, _ := New(srv.URL, "token")
	l := newArchiveLogger(t)
	a, _ := NewArchiver(c, l, true)

	if err := a.Archive("secret/data/prod/db", "archive"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArchive_WritesAndDeletes(t *testing.T) {
	wrote := false
	deleted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"data": map[string]any{"key": "val"}},
			})
		case r.Method == http.MethodPost:
			wrote = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "token")
	l := newArchiveLogger(t)
	a, _ := NewArchiver(c, l, false)

	if err := a.Archive("secret/data/prod/key", "archive"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wrote {
		t.Error("expected write to archive")
	}
	if !deleted {
		t.Error("expected delete of original")
	}
}
