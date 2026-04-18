package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newAnnotateLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newAnnotateMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			})
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestNewAnnotator_MissingClient(t *testing.T) {
	l := newAnnotateLogger(t)
	_, err := NewAnnotator(nil, l, map[string]string{"env": "prod"}, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewAnnotator_NoAnnotations(t *testing.T) {
	srv := newAnnotateMockServer(t)
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newAnnotateLogger(t)
	_, err := NewAnnotator(c, l, map[string]string{}, false)
	if err == nil {
		t.Fatal("expected error for empty annotations")
	}
}

func TestAnnotate_DryRun_NoWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			return
		}
		writes++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newAnnotateLogger(t)
	a, _ := NewAnnotator(c, l, map[string]string{"team": "platform"}, true)
	if err := a.Annotate("secret/data/app"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writes != 0 {
		t.Fatalf("expected 0 writes in dry-run, got %d", writes)
	}
}

func TestAnnotate_WritesAnnotations(t *testing.T) {
	var written map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"x": "1"}})
			return
		}
		json.NewDecoder(r.Body).Decode(&written)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newAnnotateLogger(t)
	a, _ := NewAnnotator(c, l, map[string]string{"env": "staging"}, false)
	if err := a.Annotate("secret/data/svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	anno, ok := written["_annotations"].(map[string]interface{})
	if !ok {
		t.Fatalf("_annotations not written, got: %v", written)
	}
	if anno["env"] != "staging" {
		t.Fatalf("expected env=staging, got %v", anno["env"])
	}
}
