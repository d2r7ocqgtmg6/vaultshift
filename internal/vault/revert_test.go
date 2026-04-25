package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newRevertLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newRevertMockServer(t *testing.T, written *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			*written = append(*written, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"key": "val"},
		})
	}))
}

func TestNewReverter_MissingClient(t *testing.T) {
	l := newRevertLogger(t)
	_, err := NewReverter(nil, l, false)
	if err == nil || !strings.Contains(err.Error(), "source client") {
		t.Fatalf("expected source client error, got %v", err)
	}
}

func TestNewReverter_MissingLogger(t *testing.T) {
	written := []string{}
	srv := newRevertMockServer(t, &written)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewReverter(c, nil, false)
	if err == nil || !strings.Contains(err.Error(), "logger") {
		t.Fatalf("expected logger error, got %v", err)
	}
}

func TestRevert_DryRun_NoWrite(t *testing.T) {
	written := []string{}
	srv := newRevertMockServer(t, &written)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newRevertLogger(t)
	rv, _ := NewReverter(c, l, true)
	snap := map[string]map[string]interface{}{
		"secret/a": {"k": "v"},
	}
	results := rv.Revert(snap)
	if len(written) != 0 {
		t.Fatalf("expected no writes in dry-run, got %d", len(written))
	}
	if !results[0].Skipped {
		t.Fatal("expected result to be skipped")
	}
}

func TestRevert_WritesSecrets(t *testing.T) {
	written := []string{}
	srv := newRevertMockServer(t, &written)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newRevertLogger(t)
	rv, _ := NewReverter(c, l, false)
	snap := map[string]map[string]interface{}{
		"secret/b": {"x": "y"},
	}
	results := rv.Revert(snap)
	if len(results) != 1 || !results[0].Reverted {
		t.Fatalf("expected one reverted result, got %+v", results)
	}
	if len(written) == 0 {
		t.Fatal("expected a write to have occurred")
	}
}
