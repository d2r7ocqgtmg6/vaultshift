package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newPatchLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newPatchMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			data := map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key1": "val1", "key2": "val2"}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewPatcher_MissingClient(t *testing.T) {
	l := newPatchLogger(t)
	_, err := NewPatcher(nil, l)
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewPatcher_MissingLogger(t *testing.T) {
	srv := newPatchMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewPatcher(c, nil)
	if err == nil || !strings.Contains(err.Error(), "logger") {
		t.Fatalf("expected logger error, got %v", err)
	}
}

func TestPatch_DryRun_NoWrite(t *testing.T) {
	wrote := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			wrote = true
		}
		data := map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"a": "1"}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "tok")
	l := newPatchLogger(t)
	p, _ := NewPatcher(c, l)
	p.WithPatchDryRun(true)

	res := p.Patch("secret/data/foo", map[string]string{"b": "2"})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if wrote {
		t.Fatal("dry-run should not write")
	}
	if res.Applied["b"] != "2" {
		t.Fatalf("expected applied key b=2, got %v", res.Applied)
	}
}

func TestPatch_MergesKeys(t *testing.T) {
	srv := newPatchMockServer(t)
	defer srv.Close()

	c, _ := New(srv.URL, "tok")
	l := newPatchLogger(t)
	p, _ := NewPatcher(c, l)

	res := p.Patch("secret/data/mypath", map[string]string{"key1": "updated", "key3": "new"})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Applied["key1"] != "updated" {
		t.Errorf("expected key1=updated, got %v", res.Applied["key1"])
	}
	if res.Applied["key3"] != "new" {
		t.Errorf("expected key3=new, got %v", res.Applied["key3"])
	}
}
