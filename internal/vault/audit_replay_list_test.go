package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestReplay_WritesSuccessEntries(t *testing.T) {
	var written []string
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			written = append(written, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer svr.Close()

	c, _ := New(svr.URL, "tok")
	l := newReplayLogger(t)
	rp, _ := NewReplayer(c, l, false)

	entries := []ReplayEntry{
		{Timestamp: time.Now(), Op: "write", Path: "secret/data/alpha", Status: "success", Meta: map[string]string{"key": "val"}},
		{Timestamp: time.Now(), Op: "write", Path: "secret/data/beta", Status: "success", Meta: map[string]string{"key2": "val2"}},
		{Timestamp: time.Now(), Op: "delete", Path: "secret/data/gamma", Status: "success"},
	}

	f, _ := os.CreateTemp(t.TempDir(), "audit-*.jsonl")
	enc := json.NewEncoder(f)
	for _, e := range entries {
		enc.Encode(e)
	}
	f.Close()

	n, errs := rp.Replay(f.Name())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied (only write+success), got %d", n)
	}
}

func TestReplay_InvalidFile_ReturnsError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	l := newReplayLogger(t)
	rp, _ := NewReplayer(c, l, false)

	_, errs := rp.Replay("/nonexistent/audit.jsonl")
	if len(errs) == 0 {
		t.Fatal("expected error for missing file")
	}
}
