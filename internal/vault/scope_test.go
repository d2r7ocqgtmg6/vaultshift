package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newScopeMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":{"key":"val"},"metadata":{}}}`)) 
	}))
}

func newScopeClient(addr string) *Client {
	c, _ := New(addr, "token")
	return c
}

func TestNewScoper_MissingClient(t *testing.T) {
	_, err := NewScoper(nil, "secret/myapp")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewScoper_MissingPrefix(t *testing.T) {
	srv := newScopeMockServer()
	defer srv.Close()
	c := newScopeClient(srv.URL)
	_, err := NewScoper(c, "")
	if err == nil {
		t.Fatal("expected error for empty prefix")
	}
}

func TestScoper_Enforce_Valid(t *testing.T) {
	srv := newScopeMockServer()
	defer srv.Close()
	c := newScopeClient(srv.URL)
	s, _ := NewScoper(c, "secret/myapp")

	path, err := s.Enforce("secret/myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "secret/myapp/db" {
		t.Errorf("unexpected path: %s", path)
	}
}

func TestScoper_Enforce_OutOfScope(t *testing.T) {
	srv := newScopeMockServer()
	defer srv.Close()
	c := newScopeClient(srv.URL)
	s, _ := NewScoper(c, "secret/myapp")

	_, err := s.Enforce("secret/otherapp/db")
	if err == nil {
		t.Fatal("expected error for out-of-scope path")
	}
}

func TestScoper_ScopedPaths(t *testing.T) {
	srv := newScopeMockServer()
	defer srv.Close()
	c := newScopeClient(srv.URL)
	s, _ := NewScoper(c, "secret/myapp")

	paths := []string{"secret/myapp/db", "secret/otherapp/x", "secret/myapp/redis"}
	got := s.ScopedPaths(paths)
	if len(got) != 2 {
		t.Errorf("expected 2 scoped paths, got %d", len(got))
	}
}

func TestScoper_Read_OutOfScope(t *testing.T) {
	srv := newScopeMockServer()
	defer srv.Close()
	c := newScopeClient(srv.URL)
	s, _ := NewScoper(c, "secret/myapp")

	_, err := s.Read("secret/other/secret")
	if err == nil {
		t.Fatal("expected scope error on read")
	}
}
