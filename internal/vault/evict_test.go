package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newEvictMockServer(t *testing.T, secrets map[string]interface{}, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": secrets}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				*written = d
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func newEvictLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewEvicter_MissingClient(t *testing.T) {
	_, err := NewEvicter(nil, newEvictLogger(t), []string{"secret"}, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewEvicter_NoPatterns(t *testing.T) {
	server := newEvictMockServer(t, map[string]interface{}{}, nil)
	defer server.Close()
	client, _ := New(server.URL, "token")
	_, err := NewEvicter(client, newEvictLogger(t), nil, false)
	if err == nil {
		t.Fatal("expected error for empty patterns")
	}
}

func TestEvict_DryRun_NoWrite(t *testing.T) {
	var written map[string]interface{}
	secrets := map[string]interface{}{"password": "s3cr3t", "username": "admin"}
	server := newEvictMockServer(t, secrets, &written)
	defer server.Close()

	client, _ := New(server.URL, "token")
	e, err := NewEvicter(client, newEvictLogger(t), []string{"password"}, true)
	if err != nil {
		t.Fatalf("NewEvicter: %v", err)
	}

	results, err := e.Evict("secret/data/app")
	if err != nil {
		t.Fatalf("Evict: %v", err)
	}
	if written != nil {
		t.Error("expected no write in dry-run mode")
	}
	evicted := 0
	for _, r := range results {
		if r.Evicted {
			evicted++
		}
	}
	if evicted != 1 {
		t.Errorf("expected 1 evicted key, got %d", evicted)
	}
}

func TestEvict_RemovesMatchingKeys(t *testing.T) {
	var written map[string]interface{}
	secrets := map[string]interface{}{"api_key": "abc", "host": "localhost"}
	server := newEvictMockServer(t, secrets, &written)
	defer server.Close()

	client, _ := New(server.URL, "token")
	e, err := NewEvicter(client, newEvictLogger(t), []string{"api_key"}, false)
	if err != nil {
		t.Fatalf("NewEvicter: %v", err)
	}

	_, err = e.Evict("secret/data/svc")
	if err != nil {
		t.Fatalf("Evict: %v", err)
	}
	if _, ok := written["api_key"]; ok {
		t.Error("evicted key should not appear in written data")
	}
	if _, ok := written["host"]; !ok {
		t.Error("non-evicted key should be preserved")
	}
}
