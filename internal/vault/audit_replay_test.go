package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/vaultshift/internal/audit"
)

func newReplayLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func writeReplayFile(t *testing.T, entries []ReplayEntry) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "audit-*.jsonl")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode entry: %v", err)
		}
	}
	f.Close()
	return f.Name()
}

func TestNewReplayer_MissingClient(t *testing.T) {
	l := newReplayLogger(t)
	_, err := NewReplayer(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewReplayer_MissingLogger(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	_, err := NewReplayer(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestReplay_DryRun_NoWrite(t *testing.T) {
	wrote := false
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			wrote = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	l := newReplayLogger(t)
	rp, _ := NewReplayer(c, l, true)

	entries := []ReplayEntry{
		{Timestamp: time.Now(), Op: "write", Path: "secret/foo", Status: "success"},
	}
	path := writeReplayFile(t, entries)
	n, errs := rp.Replay(path)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if n != 1 {
		t.Fatalf("expected 1 applied, got %d", n)
	}
	if wrote {
		t.Fatal("dry-run should not write")
	}
}

func TestReplay_SkipsNonWriteOps(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	l := newReplayLogger(t)
	rp, _ := NewReplayer(c, l, false)

	entries := []ReplayEntry{
		{Timestamp: time.Now(), Op: "read", Path: "secret/foo", Status: "success"},
		{Timestamp: time.Now(), Op: "write", Path: "secret/bar", Status: "error"},
	}
	path := writeReplayFile(t, entries)
	n, errs := rp.Replay(path)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if n != 0 {
		t.Fatalf("expected 0 applied, got %d", n)
	}
}
