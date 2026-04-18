package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newMaskLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func newMaskMockServer(t *testing.T, stored map[string]interface{}) (*httptest.Server, *map[string]interface{}) {
	written := &map[string]interface{}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": stored}})
			return
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		*written = body["data"].(map[string]interface{})
		w.WriteHeader(http.StatusNoContent)
	}))
	return srv, written
}

func TestNewMasker_MissingClient(t *testing.T) {
	l := newMaskLogger(t)
	_, err := NewMasker(nil, l, []string{"password"}, "***", false)
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewMasker_NoKeys(t *testing.T) {
	srv, _ := newMaskMockServer(t, nil)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newMaskLogger(t)
	_, err := NewMasker(c, l, nil, "***", false)
	if err == nil || !strings.Contains(err.Error(), "key") {
		t.Fatalf("expected key error, got %v", err)
	}
}

func TestMask_DryRun_NoWrite(t *testing.T) {
	stored := map[string]interface{}{"password": "secret", "user": "admin"}
	srv, written := newMaskMockServer(t, stored)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newMaskLogger(t)
	m, _ := NewMasker(c, l, []string{"password"}, "***", true)
	if err := m.Mask("kv/data/myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*written) != 0 {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestMask_MasksMatchingKeys(t *testing.T) {
	stored := map[string]interface{}{"password": "secret", "user": "admin"}
	srv, written := newMaskMockServer(t, stored)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newMaskLogger(t)
	m, _ := NewMasker(c, l, []string{"password"}, "REDACTED", false)
	if err := m.Mask("kv/data/myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*written)["password"] != "REDACTED" {
		t.Fatalf("expected password to be masked, got %v", (*written)["password"])
	}
	if (*written)["user"] != "admin" {
		t.Fatalf("expected user to be unchanged, got %v", (*written)["user"])
	}
}
