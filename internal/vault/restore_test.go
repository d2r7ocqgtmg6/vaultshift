package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newRestoreLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newRestoreMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestRestore_DryRun_NoWrite(t *testing.T) {
	srv := newRestoreMockServer(t)
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	logger := newRestoreLogger(t)
	restorer, err := NewRestorer(client, logger, true)
	if err != nil {
		t.Fatalf("NewRestorer: %v", err)
	}

	snap := &Snapshot{Secrets: map[string]map[string]any{
		"db/password": {"value": "secret"},
	}}

	result, err := restorer.Restore(context.Background(), snap, "dest")
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(result.Restored) != 0 {
		t.Errorf("expected no restored, got %d", len(result.Restored))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
}

func TestRestore_WritesSecrets(t *testing.T) {
	srv := newRestoreMockServer(t)
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	logger := newRestoreLogger(t)
	restorer, _ := NewRestorer(client, logger, false)

	snap := &Snapshot{Secrets: map[string]map[string]any{
		"app/key": {"value": "abc"},
	}}

	result, err := restorer.Restore(context.Background(), snap, "prod")
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(result.Restored) != 1 {
		t.Errorf("expected 1 restored, got %d", len(result.Restored))
	}
	if !strings.Contains(result.Restored[0], "prod/") {
		t.Errorf("expected dest prefix in path, got %s", result.Restored[0])
	}
}

func TestNewRestorer_MissingClient(t *testing.T) {
	logger := newRestoreLogger(t)
	_, err := NewRestorer(nil, logger, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}
