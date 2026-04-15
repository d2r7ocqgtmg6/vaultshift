package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newSnapshotMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db", "api"}},
			})
		case "/v1/secret/data/myapp/db", "/v1/secret/data/myapp/api":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "value"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newSnapshotClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestCapture_ReturnsSnapshot(t *testing.T) {
	srv := newSnapshotMockServer(t)
	defer srv.Close()

	c := newSnapshotClient(t, srv)
	ss := NewSnapshotter(c)

	snap, err := ss.Capture("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Prefix != "secret/myapp" {
		t.Errorf("expected prefix %q, got %q", "secret/myapp", snap.Prefix)
	}
	if len(snap.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(snap.Secrets))
	}
}

func TestSnapshot_SaveAndLoad(t *testing.T) {
	snap := &Snapshot{
		Prefix: "secret/test",
		Secrets: map[string]SecretData{
			"secret/test/foo": {Path: "secret/test/foo", Data: map[string]interface{}{"bar": "baz"}},
		},
	}

	tmp := filepath.Join(t.TempDir(), "snap.json")
	if err := snap.Save(tmp); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadSnapshot(tmp)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}
	if loaded.Prefix != snap.Prefix {
		t.Errorf("prefix mismatch: got %q", loaded.Prefix)
	}
	if len(loaded.Secrets) != 1 {
		t.Errorf("expected 1 secret, got %d", len(loaded.Secrets))
	}
}

func TestLoadSnapshot_InvalidFile(t *testing.T) {
	_, err := LoadSnapshot("/nonexistent/path/snap.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSnapshot_Save_InvalidPath(t *testing.T) {
	snap := &Snapshot{Prefix: "x", Secrets: map[string]SecretData{}}
	err := snap.Save("/nonexistent_dir/snap.json")
	if err == nil {
		t.Error("expected error for invalid save path")
	}
	_ = os.Remove("/nonexistent_dir/snap.json")
}
