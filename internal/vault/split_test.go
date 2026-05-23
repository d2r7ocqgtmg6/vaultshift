package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newSplitMockServer(t *testing.T, written *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"db_pass": "secret", "api_key": "key123"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			*written = append(*written, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func newSplitClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("split client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func newSplitLogger(t *testing.T) AuditLogger {
	t.Helper()
	l, err := newTestLogger()
	if err != nil {
		t.Fatalf("split logger: %v", err)
	}
	return l
}

func TestNewSplitter_MissingClient(t *testing.T) {
	_, err := NewSplitter(nil, newSplitLogger(t), map[string]string{"k": "dest"}, false)
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewSplitter_NoRoutes(t *testing.T) {
	var written []string
	srv := newSplitMockServer(t, &written)
	defer srv.Close()
	_, err := NewSplitter(newSplitClient(t, srv.URL), newSplitLogger(t), nil, false)
	if err == nil || !strings.Contains(err.Error(), "route") {
		t.Fatalf("expected route error, got %v", err)
	}
}

func TestSplit_DryRun_NoWrite(t *testing.T) {
	var written []string
	srv := newSplitMockServer(t, &written)
	defer srv.Close()
	routes := map[string]string{"db_pass": "secret/db", "api_key": "secret/api"}
	s, err := NewSplitter(newSplitClient(t, srv.URL), newSplitLogger(t), routes, true)
	if err != nil {
		t.Fatalf("new splitter: %v", err)
	}
	results, err := s.Split("secret/source")
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if len(written) != 0 {
		t.Fatalf("expected no writes in dry-run, got %d", len(written))
	}
	for _, r := range results {
		if !r.Skipped {
			t.Errorf("expected result to be skipped")
		}
	}
}

func TestSplit_WritesRoutedKeys(t *testing.T) {
	var written []string
	srv := newSplitMockServer(t, &written)
	defer srv.Close()
	routes := map[string]string{"db_pass": "secret/db"}
	s, err := NewSplitter(newSplitClient(t, srv.URL), newSplitLogger(t), routes, false)
	if err != nil {
		t.Fatalf("new splitter: %v", err)
	}
	results, err := s.Split("secret/source")
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 write, got %d", len(written))
	}
}
