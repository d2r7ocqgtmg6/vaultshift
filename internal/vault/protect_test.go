package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newProtectLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func newProtectMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]any{
		"secret/data/myapp/db": {"password": "s3cr3t"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
		case http.MethodPost, http.MethodPut:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]any); ok {
				store[path] = d
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestNewProtector_MissingClient(t *testing.T) {
	_, err := NewProtector(nil, newProtectLogger(t), false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewProtector_MissingLogger(t *testing.T) {
	srv := newProtectMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewProtector(c, nil, false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProtect_DryRun_NoWrite(t *testing.T) {
	srv := newProtectMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	p, _ := NewProtector(c, newProtectLogger(t), true)

	if err := p.Protect("secret/data/myapp/db"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ok, err := p.IsProtected("secret/data/myapp/db")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("dry-run should not have written protection flag")
	}
}

func TestProtect_WritesFlag(t *testing.T) {
	srv := newProtectMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	p, _ := NewProtector(c, newProtectLogger(t), false)

	if err := p.Protect("secret/data/myapp/db"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ok, err := p.IsProtected("secret/data/myapp/db")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected secret to be protected")
	}

	if err := p.Unprotect("secret/data/myapp/db"); err != nil {
		t.Fatalf("unprotect error: %v", err)
	}
	ok, _ = p.IsProtected("secret/data/myapp/db")
	if ok {
		t.Error("expected protection flag to be removed")
	}
}
