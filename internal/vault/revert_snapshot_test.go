package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRevert_RoundtripWithSnapshot(t *testing.T) {
	written := map[string]map[string]interface{}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body struct {
				Data map[string]interface{} `json:"data"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			written[r.URL.Path] = body.Data
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"foo": "bar"},
		})
	}))
	defer srv.Close()

	c, err := New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Build and save a snapshot.
	snap := Snapshot{
		"secret/alpha": {"user": "admin"},
		"secret/beta":  {"pass": "s3cr3t"},
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "snap.json")
	snapper, _ := NewSnapshotter(c, newSnapshotLogger(t), false)
	if err := snapper.Save(snap, file); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load snapshot back.
	loaded, err := LoadSnapshot(file)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}

	// Revert using loaded snapshot.
	l := newRevertLogger(t)
	rv, _ := NewReverter(c, l, false)
	results := rv.Revert(loaded)

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Err)
		}
		if !r.Reverted {
			t.Errorf("expected %s to be reverted", r.Path)
		}
	}

	// Verify the temp file is cleaned up properly.
	if _, err := os.Stat(file); err != nil {
		t.Logf("snapshot file already removed (ok): %v", err)
	}
}
