package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newReplicateLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newReplicateMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	written := map[string]bool{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "?list=true") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"alpha", "beta"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"value": "secret"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written[r.URL.Path] = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewReplicator_MissingClient(t *testing.T) {
	l := newReplicateLogger(t)
	_, err := NewReplicator(nil, l, false, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewReplicator_MissingLogger(t *testing.T) {
	srv := newReplicateMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	_, err := NewReplicator(c, nil, false, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestReplicate_DryRun_NoWrite(t *testing.T) {
	srv := newReplicateMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newReplicateLogger(t)
	r, _ := NewReplicator(c, l, true, false)

	results, err := r.Replicate("secret/src/", []string{"secret/dst/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, res := range results {
		if !res.Skipped {
			t.Errorf("expected all results skipped in dry-run, got written: %s", res.Path)
		}
	}
}

func TestReplicate_NoDests_ReturnsError(t *testing.T) {
	srv := newReplicateMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newReplicateLogger(t)
	r, _ := NewReplicator(c, l, false, false)

	_, err := r.Replicate("secret/src/", nil)
	if err == nil {
		t.Fatal("expected error for empty dest prefixes")
	}
}
