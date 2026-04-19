package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auditlog/logger"
)

func newNamespaceMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "LIST") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"secret-a"}},
			})
			return
		}
		if r.Method == "LIST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"secret-a"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func newNamespaceLogger(t *testing.T) *logger.Logger {
	t.Helper()
	l, err := logger.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestNewNamespaceMover_MissingSource(t *testing.T) {
	l := newNamespaceLogger(t)
	_, err := NewNamespaceMover(nil, &Client{}, l, false)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNewNamespaceMover_MissingDest(t *testing.T) {
	l := newNamespaceLogger(t)
	_, err := NewNamespaceMover(&Client{}, nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil dest")
	}
}

func TestNamespaceMover_DryRun_NoWrite(t *testing.T) {
	srv := newNamespaceMockServer()
	defer srv.Close()

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newNamespaceLogger(t)

	m, err := NewNamespaceMover(src, dst, l, true)
	if err != nil {
		t.Fatal(err)
	}

	n, err := m.Move("ns/src/", "ns/dst/")
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Error("expected at least one dry-run entry")
	}
}
