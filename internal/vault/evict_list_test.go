package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func TestEvict_IntegrationWithLister(t *testing.T) {
	writtenPaths := map[string]map[string]interface{}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list") == "true" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"alpha", "beta"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"token": "abc", "name": "svc"}},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				writtenPaths[r.URL.Path] = d
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, "token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	logger, _ := audit.New("")
	lister := NewLister(client)
	paths, err := lister.List("secret/data/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	e, err := NewEvicter(client, logger, []string{"token"}, false)
	if err != nil {
		t.Fatalf("NewEvicter: %v", err)
	}

	for _, p := range paths {
		if _, err := e.Evict(p); err != nil {
			t.Errorf("Evict(%s): %v", p, err)
		}
	}

	if len(writtenPaths) != len(paths) {
		t.Errorf("expected %d writes, got %d", len(paths), len(writtenPaths))
	}
	for _, d := range writtenPaths {
		if _, ok := d["token"]; ok {
			t.Error("evicted key 'token' should not appear in written data")
		}
		if _, ok := d["name"]; !ok {
			t.Error("non-evicted key 'name' should be preserved")
		}
	}
}
