package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newPruneLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newPruneMockServer(secrets map[string]map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			path := strings.TrimPrefix(r.URL.Path, "/v1/")
			delete(secrets, path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if strings.Contains(r.URL.RawQuery, "list=true") {
			keys := []string{}
			prefix := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/v1/")
			for k := range secrets {
				if strings.HasPrefix(k, prefix) {
					keys = append(keys, strings.TrimPrefix(k, prefix+"/"))
				}
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": keys}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestNewPruner_MissingClient(t *testing.T) {
	l := newPruneLogger(t)
	_, err := NewPruner(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestPrune_DryRun_NoDelete(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/app/a": {"val": "1"},
		"secret/app/b": {"val": "2"},
	}
	srv := newPruneMockServer(secrets)
	defer srv.Close()

	c, _ := New(srv.URL, "token")
	l := newPruneLogger(t)
	p, _ := NewPruner(c, l, true)

	result, err := p.Prune("secret/app", []string{"secret/app/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 0 {
		t.Errorf("expected no deletions in dry-run, got %v", result.Deleted)
	}
}

func TestPrune_DeletesOrphaned(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/app/a": {"val": "1"},
		"secret/app/b": {"val": "2"},
		"secret/app/c": {"val": "3"},
	}
	srv := newPruneMockServer(secrets)
	defer srv.Close()

	c, _ := New(srv.URL, "token")
	l := newPruneLogger(t)
	p, _ := NewPruner(c, l, false)

	result, err := p.Prune("secret/app", []string{"secret/app/a", "secret/app/c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(result.Deleted)
	if len(result.Deleted) != 1 || result.Deleted[0] != "secret/app/b" {
		t.Errorf("expected [secret/app/b] deleted, got %v", result.Deleted)
	}
}
