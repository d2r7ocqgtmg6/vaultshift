package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultshift/internal/audit"
)

func newSanitizeLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, _ := audit.New(audit.Config{})
	return l
}

func newSanitizeMockServer(t *testing.T, initial map[string]any) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	writes := &[]map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"data": initial})
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		*writes = append(*writes, body)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	return srv, writes
}

func TestNewSanitizer_MissingClient(t *testing.T) {
	_, err := NewSanitizer(SanitizerConfig{Keys: []string{"password"}})
	if err == nil {
		t.Fatal("expected error for missing client")
	}
}

func TestNewSanitizer_NoKeys(t *testing.T) {
	srv, _ := newSanitizeMockServer(t, nil)
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	_, err := NewSanitizer(Sanitin	if err == nil {
		t.Fatal("expected error for empty keys")
	}
}

func TestSanitize_DryRun_NoWrite(t *testing.T) {
	initial := map[string]any{"password": "secret", "user": "admin"}
	srv, writes := newSanitizeMockServer(t, initial)
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	s, _ := NewSanitizer(SanitizerConfig{
		Client: c, Logger: newSanitizeLogger(t),
		Keys: []string{"password"}, DryRun: true,
	})
	modified, err := s.Sanitize("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modified {
		t.Error("expected modified=true")
	}
	if len(*writes) != 0 {
		t.Errorf("expected no writes in dry-run, got %d", len(*writes))
	}
}

func TestSanitize_RemovesMatchingKey(t *testing.T) {
	initial := map[string]any{"password": "secret", "user": "admin"}
	srv, writes := newSanitizeMockServer(t, initial)
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	s, _ := NewSanitizer(SanitizerConfig{
		Client: c, Logger: newSanitizeLogger(t),
		Keys: []string{"password"}, DryRun: false,
	})
	modified, err := s.Sanitize("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modified {
		t.Error("expected modified=true")
	}
	if len(*writes) != 1 {
		t.Errorf("expected 1 write, got %d", len(*writes))
	}
}

func TestSanitize_NoMatch_NotModified(t *testing.T) {
	initial := map[string]any{"user": "admin"}
	srv, _ := newSanitizeMockServer(t, initial)
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	s, _ := NewSanitizer(SanitizerConfig{
		Client: c, Keys: []string{"password"}, DryRun: false,
	})
	modified, err := s.Sanitize("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if modified {
		t.Error("expected modified=false")
	}
}
