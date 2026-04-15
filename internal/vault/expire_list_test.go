package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Integration-style test: Lister + Expirer work together to purge expired
// secrets discovered under a prefix.
func TestExpire_IntegrationWithLister(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	future := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	deleted := &[]string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			*deleted = append(*deleted, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/v1/secret/metadata/env" && r.URL.Query().Get("list") == "true":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"old", "new"}},
			})
		case r.URL.Path == "/v1/secret/data/env/old":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"expires_at": past}},
			})
		case r.URL.Path == "/v1/secret/data/env/new":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"expires_at": future}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}

	lister := NewLister(client)
	paths, err := lister.List("secret/metadata/env")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	logger, _ := audit.New("")
	expirer := NewExpirer(client, logger, "expires_at", false)
	results := expirer.CheckAndPurge(paths)

	expiredCount := 0
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Error)
		}
		if r.Expired {
			expiredCount++
		}
	}
	if expiredCount != 1 {
		t.Fatalf("expected 1 expired secret, got %d", expiredCount)
	}
	if len(*deleted) != 1 {
		t.Fatalf("expected 1 deletion, got %v", *deleted)
	}
}
