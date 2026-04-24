package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dreamsofcode-io/vaultshift/internal/audit"
)

func newPivotLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newPivotMockServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			for path, data := range secrets {
				if r.URL.Path == "/v1/"+path {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewPivoter_MissingSource(t *testing.T) {
	l := newPivotLogger(t)
	_, err := NewPivoter(nil, &Client{}, l, false)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNewPivoter_MissingDest(t *testing.T) {
	l := newPivotLogger(t)
	_, err := NewPivoter(&Client{}, nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil dest")
	}
}

func TestPivot_DryRun_NoWrite(t *testing.T) {
	written := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written++
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"k": "v"}})
	}))
	t.Cleanup(srv.Close)

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newPivotLogger(t)
	p, err := NewPivoter(src, dst, l, true)
	if err != nil {
		t.Fatalf("NewPivoter: %v", err)
	}
	if err := p.Pivot([]string{"secret/foo"}); err != nil {
		t.Fatalf("Pivot: %v", err)
	}
	if written != 0 {
		t.Errorf("expected 0 writes in dry-run, got %d", written)
	}
}

func TestPivot_WritesSecrets(t *testing.T) {
	written := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written++
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"k": "v"}})
	}))
	t.Cleanup(srv.Close)

	src, _ := New(srv.URL, "tok")
	dst, _ := New(srv.URL, "tok")
	l := newPivotLogger(t)
	p, err := NewPivoter(src, dst, l, false)
	if err != nil {
		t.Fatalf("NewPivoter: %v", err)
	}
	if err := p.Pivot([]string{"secret/a", "secret/b"}); err != nil {
		t.Fatalf("Pivot: %v", err)
	}
	if written != 2 {
		t.Errorf("expected 2 writes, got %d", written)
	}
}
