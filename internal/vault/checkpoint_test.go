package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newCheckpointMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/app":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/app/db":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newCheckpointClient(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := New(addr, "token")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewCheckpointer_MissingClient(t *testing.T) {
	_, err := NewCheckpointer(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckpoint_DryRun_NoFile(t *testing.T) {
	srv := newCheckpointMockServer()
	defer srv.Close()
	c := newCheckpointClient(t, srv.URL)
	cp, _ := NewCheckpointer(c)

	path := filepath.Join(t.TempDir(), "cp.json")
	if err := cp.Save("secret/app", path, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("file should not exist in dry-run")
	}
}

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	srv := newCheckpointMockServer()
	defer srv.Close()
	c := newCheckpointClient(t, srv.URL)
	cp, _ := NewCheckpointer(c)

	path := filepath.Join(t.TempDir(), "cp.json")
	if err := cp.Save("secret/app", path, false); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadCheckpoint(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Prefix != "secret/app" {
		t.Errorf("expected prefix secret/app, got %s", loaded.Prefix)
	}
}

func TestLoadCheckpoint_InvalidFile(t *testing.T) {
	_, err := LoadCheckpoint("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error")
	}
}
