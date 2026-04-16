package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newCopyLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newCopyMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"data":{"key":"val"},"metadata":{}}}`))  
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func newCopyClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewCopier_MissingClient(t *testing.T) {
	l := newCopyLogger(t)
	if _, err := NewCopier(nil, nil, l, false); err == nil {
		t.Fatal("expected error for nil src")
	}
}

func TestCopy_DryRun_NoWrite(t *testing.T) {
	srv := newCopyMockServer(t)
	defer srv.Close()
	client := newCopyClient(t, srv)
	logger := newCopyLogger(t)

	cp, err := NewCopier(client, client, logger, true)
	if err != nil {
		t.Fatalf("NewCopier: %v", err)
	}

	if err := cp.Copy("secret/data/src", "secret/data/dst"); err != nil {
		t.Fatalf("Copy dry-run: %v", err)
	}
}

func TestCopy_WritesSecret(t *testing.T) {
	srv := newCopyMockServer(t)
	defer srv.Close()
	client := newCopyClient(t, srv)
	logger := newCopyLogger(t)

	cp, err := NewCopier(client, client, logger, false)
	if err != nil {
		t.Fatalf("NewCopier: %v", err)
	}

	if err := cp.Copy("secret/data/src", "secret/data/dst"); err != nil {
		t.Fatalf("Copy: %v", err)
	}
}
