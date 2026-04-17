package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newPinLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func newPinMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewPinner_MissingClient(t *testing.T) {
	l := newPinLogger(t)
	_, err := NewPinner(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestPin_DryRun_NoWrite(t *testing.T) {
	srv := newPinMockServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	if err != nil {
		t.Fatal(err)
	}
	l := newPinLogger(t)
	pinner, err := NewPinner(client, l, true)
	if err != nil {
		t.Fatal(err)
	}

	res := pinner.Pin("secret/data/myapp")
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.Pinned {
		t.Error("dry run should not pin")
	}
}

func TestPin_WritesPin(t *testing.T) {
	written := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client, err := New(srv.URL, "token")
	if err != nil {
		t.Fatal(err)
	}
	l := newPinLogger(t)
	pinner, _ := NewPinner(client, l, false)
	res := pinner.Pin("secret/data/myapp")
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if !written {
		t.Error("expected write to be called")
	}
}
