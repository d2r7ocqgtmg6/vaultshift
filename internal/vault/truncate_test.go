package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newTruncateLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newTruncateMockServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewTruncater_MissingClient(t *testing.T) {
	l := newTruncateLogger(t)
	_, err := NewTruncater(nil, l, 10, []string{"key"}, false)
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewTruncater_MissingLogger(t *testing.T) {
	srv := newTruncateMockServer(t, nil)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewTruncater(c, nil, 10, []string{"key"}, false)
	if err == nil || !strings.Contains(err.Error(), "logger") {
		t.Fatalf("expected logger error, got %v", err)
	}
}

func TestNewTruncater_InvalidMaxLen(t *testing.T) {
	srv := newTruncateMockServer(t, nil)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTruncateLogger(t)
	_, err := NewTruncater(c, l, 0, []string{"key"}, false)
	if err == nil || !strings.Contains(err.Error(), "maxLen") {
		t.Fatalf("expected maxLen error, got %v", err)
	}
}

func TestTruncate_DryRun_NoWrite(t *testing.T) {
	wrote := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": map[string]interface{}{"token": "supersecretvalue"}}})
			return
		}
		wrote = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTruncateLogger(t)
	tr, err := NewTruncater(c, l, 5, []string{"token"}, true)
	if err != nil {
		t.Fatalf("NewTruncater: %v", err)
	}
	if err := tr.Truncate("secret/data/app"); err != nil {
		t.Fatalf("Truncate: %v", err)
	}
	if wrote {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestTruncate_WritesWhenExceedsMaxLen(t *testing.T) {
	var written map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": map[string]interface{}{"pw": "toolongpassword"}}})
			return
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		written = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTruncateLogger(t)
	tr, err := NewTruncater(c, l, 4, []string{"pw"}, false)
	if err != nil {
		t.Fatalf("NewTruncater: %v", err)
	}
	if err := tr.Truncate("secret/data/app"); err != nil {
		t.Fatalf("Truncate: %v", err)
	}
	if written == nil {
		t.Fatal("expected write to be called")
	}
}
