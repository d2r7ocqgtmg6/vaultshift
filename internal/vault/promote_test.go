package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPromoteMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/staging/":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"keys":["db","api"]}}`)) //nolint:errcheck
		case r.Method == http.MethodGet && (r.URL.Path == "/v1/secret/data/staging/db" || r.URL.Path == "/v1/secret/data/staging/api"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"data":{"key":"value"}}}`)) //nolint:errcheck
		case r.Method == http.MethodGet && (r.URL.Path == "/v1/secret/data/prod/db" || r.URL.Path == "/v1/secret/data/prod/api"):
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut || r.Method == http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newPromoteClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c
}

func TestPromote_DryRun_NoWrite(t *testing.T) {
	srv := newPromoteMockServer()
	defer srv.Close()

	client := newPromoteClient(t, srv)
	logger := newTestLogger()
	p := NewPromoter(client, client, logger, true)

	results, err := p.Promote("secret/staging/", "secret/prod/", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	for _, r := range results {
		if !r.DryRun {
			t.Errorf("expected dry_run=true for %q", r.Path)
		}
		if r.Err != nil {
			t.Errorf("unexpected error for %q: %v", r.Path, r.Err)
		}
	}
}

func TestPromote_SkipExisting(t *testing.T) {
	skipSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/staging/":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"keys":["db"]}}`)) //nolint:errcheck
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"data":{"key":"value"}}}`)) //nolint:errcheck
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer skipSrv.Close()

	client := newPromoteClient(t, skipSrv)
	logger := newTestLogger()
	p := NewPromoter(client, client, logger, false)

	results, err := p.Promote("secret/staging/", "secret/prod/", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Skipped {
			t.Errorf("expected skipped=true for %q", r.Path)
		}
	}
}
