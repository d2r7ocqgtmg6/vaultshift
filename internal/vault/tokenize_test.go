package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultshift/internal/audit"
)

func newTokenizeLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newTokenizeMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t", "user": "admin"}},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewTokenizer_MissingClient(t *testing.T) {
	l := newTokenizeLogger(t)
	_, err := NewTokenizer(nil, l, "tokens", false)
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewTokenizer_MissingLogger(t *testing.T) {
	srv := newTokenizeMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewTokenizer(c, nil, "tokens", false)
	if err == nil || !strings.Contains(err.Error(), "logger") {
		t.Fatalf("expected logger error, got %v", err)
	}
}

func TestNewTokenizer_MissingTokenNS(t *testing.T) {
	srv := newTokenizeMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTokenizeLogger(t)
	_, err := NewTokenizer(c, l, "  ", false)
	if err == nil || !strings.Contains(err.Error(), "tokenNS") {
		t.Fatalf("expected tokenNS error, got %v", err)
	}
}

func TestTokenize_DryRun_NoWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t"}},
		})
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTokenizeLogger(t)
	tz, _ := NewTokenizer(c, l, "tokens", true)
	tokens, err := tz.Tokenize("secret/app", []string{"password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writes != 0 {
		t.Errorf("expected no writes in dry-run, got %d", writes)
	}
	if _, ok := tokens["password"]; !ok {
		t.Error("expected token for password key")
	}
}

func TestTokenize_NoKeys_ReturnsError(t *testing.T) {
	srv := newTokenizeMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTokenizeLogger(t)
	tz, _ := NewTokenizer(c, l, "tokens", false)
	_, err := tz.Tokenize("secret/app", []string{})
	if err == nil {
		t.Fatal("expected error for empty keys")
	}
}
