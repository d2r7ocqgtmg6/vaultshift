package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newCloneMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "list") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"alpha", "beta"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"value": "secret123"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func newCloneClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("newCloneClient: %v", err)
	}
	return c
}

func newCloneLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("newCloneLogger: %v", err)
	}
	return l
}

func TestClone_DryRun_NoWrite(t *testing.T) {
	srv := newCloneMockServer()
	defer srv.Close()

	client := newCloneClient(t, srv)
	logger := newCloneLogger(t)
	cloner := NewCloner(client, logger, true)

	result, err := cloner.Clone("secret/src/", "secret/dest/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Copied) != 0 {
		t.Errorf("expected no copies in dry-run, got %d", len(result.Copied))
	}
	if len(result.Skipped) == 0 {
		t.Error("expected skipped entries in dry-run")
	}
}

func TestClone_WritesSecrets(t *testing.T) {
	srv := newCloneMockServer()
	defer srv.Close()

	client := newCloneClient(t, srv)
	logger := newCloneLogger(t)
	cloner := NewCloner(client, logger, false)

	result, err := cloner.Clone("secret/src/", "secret/dest/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Copied) == 0 {
		t.Error("expected copied entries")
	}
}
