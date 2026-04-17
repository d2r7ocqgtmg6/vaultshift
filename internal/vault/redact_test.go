package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newRedactMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"password": "s3cr3t",
						"username": "admin",
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func newRedactClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newRedactLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestNewRedactor_MissingClient(t *testing.T) {
	_, err := NewRedactor(nil, nil, []string{"password"}, "")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewRedactor_NoKeys(t *testing.T) {
	srv := newRedactMockServer(t)
	defer srv.Close()
	c := newRedactClient(t, srv)
	_, err := NewRedactor(c, nil, nil, "")
	if err == nil {
		t.Fatal("expected error for empty keys")
	}
}

func TestRedact_DryRun_NoWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t"}},
		})
	}))
	defer srv.Close()
	c := newRedactClient(t, srv)
	l := newRedactLogger(t)
	r, err := NewRedactor(c, l, []string{"password"}, "REDACTED")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Redact("secret/data/myapp", true); err != nil {
		t.Fatal(err)
	}
	if writes != 0 {
		t.Errorf("expected 0 writes in dry-run, got %d", writes)
	}
}

func TestRedact_WritesRedactedSecret(t *testing.T) {
	written := map[string]interface{}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&written)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t", "user": "admin"}},
		})
	}))
	defer srv.Close()
	c := newRedactClient(t, srv)
	r, err := NewRedactor(c, nil, []string{"password"}, "***")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Redact("secret/data/myapp", false); err != nil {
		t.Fatal(err)
	}
}
