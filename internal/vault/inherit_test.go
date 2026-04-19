package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newInheritLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newInheritMockServer(parent, child map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var data map[string]interface{}
			if r.URL.Path == "/v1/secret/data/parent" {
				data = parent
			} else {
				data = child
			}
			if data == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewInheritor_MissingClient(t *testing.T) {
	_, err := NewInheritor(nil, newInheritLogger(t), false, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestInherit_DryRun_NoWrite(t *testing.T) {
	srv := newInheritMockServer(
		map[string]interface{}{"db_pass": "secret"},
		map[string]interface{}{"app_key": "val"},
	)
	defer srv.Close()

	c, _ := New(srv.URL, "token")
	h, _ := NewInheritor(c, newInheritLogger(t), true, false)

	n, err := h.Inherit("secret/data/parent", "secret/data/child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 inherited key, got %d", n)
	}
}

func TestInherit_SkipsExistingWithoutOverwrite(t *testing.T) {
	srv := newInheritMockServer(
		map[string]interface{}{"key": "parent_val"},
		map[string]interface{}{"key": "child_val"},
	)
	defer srv.Close()

	c, _ := New(srv.URL, "token")
	h, _ := NewInheritor(c, newInheritLogger(t), false, false)

	n, err := h.Inherit("secret/data/parent", "secret/data/child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 inherited keys, got %d", n)
	}
}
