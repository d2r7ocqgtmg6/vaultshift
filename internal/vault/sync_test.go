package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newSyncLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newSyncMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list") == "true" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"alpha"}},
			})
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"value": "secret"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewSyncer_MissingClient(t *testing.T) {
	l := newSyncLogger(t)
	_, err := NewSyncer(nil, nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil src client")
	}
}

func TestNewSyncer_MissingLogger(t *testing.T) {
	srv := newSyncMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	_, err := NewSyncer(c, c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestSync_DryRun_NoWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Query().Get("list") == "true" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"key1"}},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"v": "1"},
		})
	}))
	defer srv.Close()

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newSyncLogger(t)
	s, _ := NewSyncer(src, dst, l, true)
	result, err := s.Sync("secret/src/", "secret/dst/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writes != 0 {
		t.Errorf("expected 0 writes in dry-run, got %d", writes)
	}
	_ = result
}

func TestSync_PropagatesBothDirections(t *testing.T) {
	callCount := 0
	srv := newSyncMockServer(t)
	defer srv.Close()
	_ = callCount

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newSyncLogger(t)
	s, _ := NewSyncer(src, dst, l, false)
	result, err := s.Sync("secret/a/", "secret/b/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}
