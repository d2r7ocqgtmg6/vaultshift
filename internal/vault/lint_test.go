package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newLintMockServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if data == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func newLintLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newLintClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewLinter_MissingClient(t *testing.T) {
	l := newLintLogger(t)
	_, err := NewLinter(nil, l, NoEmptyKeys)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewLinter_MissingLogger(t *testing.T) {
	srv := newLintMockServer(map[string]interface{}{"k": "v"})
	defer srv.Close()
	c := newLintClient(t, srv)
	_, err := NewLinter(c, nil, NoEmptyKeys)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewLinter_NoRules(t *testing.T) {
	srv := newLintMockServer(map[string]interface{}{"k": "v"})
	defer srv.Close()
	c := newLintClient(t, srv)
	l := newLintLogger(t)
	_, err := NewLinter(c, l)
	if err == nil {
		t.Fatal("expected error for no rules")
	}
}

func TestLint_NoViolations(t *testing.T) {
	srv := newLintMockServer(map[string]interface{}{"api_key": "abc123"})
	defer srv.Close()
	c := newLintClient(t, srv)
	l := newLintLogger(t)
	linter, err := NewLinter(c, l, NoEmptyKeys, NoUpperCaseKeys)
	if err != nil {
		t.Fatalf("NewLinter: %v", err)
	}
	result, err := linter.Lint("secret/data/app")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(result.Violations) != 0 {
		t.Fatalf("expected no violations, got %v", result.Violations)
	}
}

func TestLint_DetectsEmptyValue(t *testing.T) {
	srv := newLintMockServer(map[string]interface{}{"password": ""})
	defer srv.Close()
	c := newLintClient(t, srv)
	l := newLintLogger(t)
	linter, _ := NewLinter(c, l, NoEmptyKeys)
	result, err := linter.Lint("secret/data/app")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(result.Violations) == 0 {
		t.Fatal("expected violation for empty value")
	}
}

func TestLint_DetectsUpperCaseKey(t *testing.T) {
	srv := newLintMockServer(map[string]interface{}{"ApiKey": "val"})
	defer srv.Close()
	c := newLintClient(t, srv)
	l := newLintLogger(t)
	linter, _ := NewLinter(c, l, NoUpperCaseKeys)
	result, err := linter.Lint("secret/data/app")
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(result.Violations) == 0 {
		t.Fatal("expected violation for uppercase key")
	}
}
