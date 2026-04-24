package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultshift/internal/audit"
)

func newFlattenLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newFlattenMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/src":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"a", "b"}},
			})
		case "/v1/secret/data/src/a":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key1": "val1"}},
			})
		case "/v1/secret/data/src/b":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key2": "val2"}},
			})
		case "/v1/secret/data/dst/merged":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestNewFlattener_MissingClient(t *testing.T) {
	_, err := NewFlattener(nil, newFlattenLogger(t), false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewFlattener_MissingLogger(t *testing.T) {
	svr := newFlattenMockServer(t)
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	_, err := NewFlattener(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestFlatten_DryRun_NoWrite(t *testing.T) {
	wrote := false
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			wrote = true
		}
		switch r.URL.Path {
		case "/v1/secret/metadata/src":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"x"}},
			})
		case "/v1/secret/data/src/x":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"k": "v"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer svr.Close()

	c, _ := New(svr.URL, "tok")
	f, _ := NewFlattener(c, newFlattenLogger(t), true)
	if err := f.Flatten("src", "dst/merged"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wrote {
		t.Error("dry-run should not write")
	}
}

func TestFlatten_MergesAndWrites(t *testing.T) {
	svr := newFlattenMockServer(t)
	defer svr.Close()

	c, _ := New(svr.URL, "tok")
	f, _ := NewFlattener(c, newFlattenLogger(t), false)
	if err := f.Flatten("src", "dst/merged"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
