package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newMirrorLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newMirrorMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "?list=true") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"alpha", "beta"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"value": "secret"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewMirrorer_MissingClient(t *testing.T) {
	l := newMirrorLogger(t)
	_, err := NewMirrorer(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewMirrorer_MissingLogger(t *testing.T) {
	srv := newMirrorMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewMirrorer(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestMirror_DryRun_NoWrite(t *testing.T) {
	srv := newMirrorMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newMirrorLogger(t)
	m, err := NewMirrorer(c, l, true)
	if err != nil {
		t.Fatalf("NewMirrorer: %v", err)
	}
	results, err := m.Mirror("src/", "dst/")
	if err != nil {
		t.Fatalf("Mirror: %v", err)
	}
	for _, r := range results {
		if !r.Skipped {
			t.Errorf("expected skipped=true for %s", r.Path)
		}
	}
}

func TestMirror_WritesSecrets(t *testing.T) {
	wrote := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "list=true") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"key1"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"val": "x"},
			})
			return
		}
		wrote[r.URL.Path] = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newMirrorLogger(t)
	m, _ := NewMirrorer(c, l, false)
	results, err := m.Mirror("src/", "dst/")
	if err != nil {
		t.Fatalf("Mirror: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Error)
		}
	}
}
