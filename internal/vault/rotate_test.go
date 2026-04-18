package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/subtlepseudonym/vaultshift/internal/audit"
)

func newRotateLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func newRotateMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"apikey": "old-value"}},
			})
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestNewRotator_MissingClient(t *testing.T) {
	l := newRotateLogger(t)
	_, err := NewRotator(nil, l)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewRotator_MissingLogger(t *testing.T) {
	srv := newRotateMockServer(t)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	_, err := NewRotator(c, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestRotate_DryRun_NoWrite(t *testing.T) {
	wrote := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			wrote = true
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"k": "v"}},
		})
	}))
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	l := newRotateLogger(t)
	r, _ := NewRotator(c, l, WithDryRun(true))
	result, err := r.Rotate("secret/data/test")
	if err != nil {
		t.Fatal(err)
	}
	if wrote {
		t.Error("dry-run should not write")
	}
	if result["k"] != "v_rotated" {
		t.Errorf("unexpected value: %s", result["k"])
	}
}

func TestRotate_WritesSecret(t *testing.T) {
	wrote := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			wrote = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"k": "v"}},
		})
	}))
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "t"})
	l := newRotateLogger(t)
	r, _ := NewRotator(c, l)
	_, err := r.Rotate("secret/data/test")
	if err != nil {
		t.Fatal(err)
	}
	if !wrote {
		t.Error("expected write")
	}
}
