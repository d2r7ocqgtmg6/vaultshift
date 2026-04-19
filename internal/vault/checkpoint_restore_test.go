package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
)

func TestLoadCheckpoint_RestoreRoundtrip(t *testing.T) {
	wrote := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/ns":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"key1"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/ns/key1":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"val": "abc"}},
			})
		case r.Method == http.MethodPut || r.Method == http.MethodPost:
			wrote[r.URL.Path] = "ok"
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	cp, _ := NewCheckpointer(client)

	path := filepath.Join(t.TempDir(), "cp.json")
	if err := cp.Save("secret/ns", path, false); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadCheckpoint(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if loaded.CreatedAt.After(time.Now().Add(time.Second)) {
		t.Error("CreatedAt is in the future")
	}
	if _, ok := loaded.Secrets["secret/ns/key1"]; !ok {
		t.Errorf("expected secret/ns/key1 in checkpoint, got keys: %v", keysOf(loaded.Secrets))
	}
}

func keysOf(m map[string]api.Secret) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
