package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newShrinkLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newShrinkMockServer(t *testing.T, secret map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": secret},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewShrinker_MissingClient(t *testing.T) {
	l := newShrinkLogger(t)
	_, err := NewShrinker(nil, l, 10, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewShrinker_MissingLogger(t *testing.T) {
	srv := newShrinkMockServer(t, map[string]interface{}{})
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	_, err := NewShrinker(c, nil, 10, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewShrinker_InvalidMaxLen(t *testing.T) {
	srv := newShrinkMockServer(t, map[string]interface{}{})
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newShrinkLogger(t)
	_, err := NewShrinker(c, l, 0, false)
	if err == nil {
		t.Fatal("expected error for maxLen=0")
	}
}

func TestShrink_DryRun_NoWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"big": "aaaaabbbbbccccc"}},
		})
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newShrinkLogger(t)
	sh, _ := NewShrinker(c, l, 5, true)
	removed, err := sh.Shrink("secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 1 || removed[0] != "big" {
		t.Fatalf("expected [big] removed, got %v", removed)
	}
	if writes != 0 {
		t.Fatalf("expected no writes in dry-run, got %d", writes)
	}
}

func TestShrink_RemovesOversizedKeys(t *testing.T) {
	var written map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&written)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"small": "hi", "large": "verylongvalue"}},
		})
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newShrinkLogger(t)
	sh, _ := NewShrinker(c, l, 5, false)
	removed, err := sh.Shrink("secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 1 || removed[0] != "large" {
		t.Fatalf("expected [large] removed, got %v", removed)
	}
}
