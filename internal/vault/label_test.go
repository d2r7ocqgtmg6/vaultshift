package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newLabelLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newLabelMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			payload := map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			}
			_ = json.NewEncoder(w).Encode(payload)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewLabeler_MissingClient(t *testing.T) {
	_, err := NewLabeler(nil, newLabelLogger(t), map[string]string{"env": "prod"}, nil, false)
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewLabeler_NoLabels(t *testing.T) {
	srv := newLabelMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	_, err := NewLabeler(c, newLabelLogger(t), nil, nil, false)
	if err == nil {
		t.Fatal("expected error for no labels")
	}
}

func TestLabel_DryRun_NoWrite(t *testing.T) {
	wrote := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			wrote = true
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"x": "y"}})
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l, _ := NewLabeler(c, newLabelLogger(t), map[string]string{"env": "staging"}, nil, true)
	if err := l.Label([]string{"secret/app"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wrote {
		t.Fatal("dry-run should not write")
	}
}

func TestLabel_WritesLabels(t *testing.T) {
	var written map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&written)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"key": "val"}})
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l, _ := NewLabeler(c, newLabelLogger(t), map[string]string{"team": "ops"}, nil, false)
	if err := l.Label([]string{"secret/svc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written == nil {
		t.Fatal("expected write to occur")
	}
}
