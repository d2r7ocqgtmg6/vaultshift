package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newDrainLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newDrainMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "list"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"alpha"}},
			})
		case r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}},
			})
		case r.Method == http.MethodPost, r.Method == http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestNewDrainer_MissingClient(t *testing.T) {
	l := newDrainLogger(t)
	if _, err := NewDrainer(nil, nil, l); err == nil {
		t.Fatal("expected error for nil src client")
	}
}

func TestDrain_DryRun_NoDelete(t *testing.T) {
	srv := newDrainMockServer(t)
	defer srv.Close()

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newDrainLogger(t)

	dr, err := NewDrainer(src, dst, l, WithDrainDryRun(true))
	if err != nil {
		t.Fatalf("NewDrainer: %v", err)
	}

	n, err := dr.Drain("secret/src/", "secret/dst/")
	if err != nil {
		t.Fatalf("Drain: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 moved, got %d", n)
	}
}

func TestDrain_MovesAndDeletes(t *testing.T) {
	srv := newDrainMockServer(t)
	defer srv.Close()

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newDrainLogger(t)

	dr, err := NewDrainer(src, dst, l)
	if err != nil {
		t.Fatalf("NewDrainer: %v", err)
	}

	n, err := dr.Drain("secret/src/", "secret/dst/")
	if err != nil {
		t.Fatalf("Drain: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 moved, got %d", n)
	}
}
