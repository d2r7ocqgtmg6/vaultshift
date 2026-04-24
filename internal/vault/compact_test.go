package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newCompactLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newCompactMockServer(secrets map[string]map[string]interface{}, deleted *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			for k, v := range secrets {
				if "/v1/"+k == path {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{"data": v})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodDelete:
			for k := range secrets {
				if "/v1/"+k == path {
					*deleted = append(*deleted, k)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestNewCompacter_MissingClient(t *testing.T) {
	_, err := NewCompacter(nil, newCompactLogger(t), false, true, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewCompacter_NoMode(t *testing.T) {
	cl := &Client{}
	_, err := NewCompacter(cl, newCompactLogger(t), false, false, false)
	if err == nil {
		t.Fatal("expected error when neither removeEmpty nor removeBlank is set")
	}
}

func TestCompact_DryRun_NoDelete(t *testing.T) {
	deleted := []string{}
	srv := newCompactMockServer(map[string]map[string]interface{}{
		"secret/empty": {},
	}, &deleted)
	defer srv.Close()

	cl, _ := New(srv.URL, "tok")
	comp, err := NewCompacter(cl, newCompactLogger(t), true, true, false)
	if err != nil {
		t.Fatalf("NewCompacter: %v", err)
	}

	removed, err := comp.Compact([]string{"secret/empty"})
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed path, got %d", len(removed))
	}
	if len(deleted) != 0 {
		t.Fatalf("dry-run should not delete, got %v", deleted)
	}
}

func TestCompact_DeletesEmptySecret(t *testing.T) {
	deleted := []string{}
	srv := newCompactMockServer(map[string]map[string]interface{}{
		"secret/empty": {},
		"secret/full":  {"key": "value"},
	}, &deleted)
	defer srv.Close()

	cl, _ := New(srv.URL, "tok")
	comp, _ := NewCompacter(cl, newCompactLogger(t), false, true, false)

	removed, _ := comp.Compact([]string{"secret/empty", "secret/full"})
	if len(removed) != 1 || removed[0] != "secret/empty" {
		t.Fatalf("expected only secret/empty removed, got %v", removed)
	}
	if len(deleted) != 1 || deleted[0] != "secret/empty" {
		t.Fatalf("expected secret/empty deleted, got %v", deleted)
	}
}

func TestCompact_DeletesAllBlankSecret(t *testing.T) {
	deleted := []string{}
	srv := newCompactMockServer(map[string]map[string]interface{}{
		"secret/blank": {"k": ""},
		"secret/mixed": {"k": "v"},
	}, &deleted)
	defer srv.Close()

	cl, _ := New(srv.URL, "tok")
	comp, _ := NewCompacter(cl, newCompactLogger(t), false, false, true)

	removed, _ := comp.Compact([]string{"secret/blank", "secret/mixed"})
	if len(removed) != 1 || removed[0] != "secret/blank" {
		t.Fatalf("expected only secret/blank removed, got %v", removed)
	}
}
