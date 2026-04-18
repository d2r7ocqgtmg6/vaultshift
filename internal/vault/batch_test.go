package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newBatchLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newBatchMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut, http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestNewBatcher_MissingClient(t *testing.T) {
	l := newBatchLogger(t)
	_, err := NewBatcher(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewBatcher_MissingLogger(t *testing.T) {
	srv := newBatchMockServer()
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	_, err := NewBatcher(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestBatch_WriteAll_DryRun_NoWrite(t *testing.T) {
	srv := newBatchMockServer()
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newBatchLogger(t)
	b, _ := NewBatcher(c, l, true)

	results := b.WriteAll(map[string]map[string]interface{}{
		"secret/a": {"key": "val"},
		"secret/b": {"key": "val2"},
	})
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for %s", r.Path)
		}
	}
}

func TestBatch_DeleteAll_DryRun_NoDelete(t *testing.T) {
	srv := newBatchMockServer()
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newBatchLogger(t)
	b, _ := NewBatcher(c, l, true)

	results := b.DeleteAll([]string{"secret/a", "secret/b"})
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for %s", r.Path)
		}
	}
}

func TestBatch_WriteAll_Live(t *testing.T) {
	srv := newBatchMockServer()
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	l := newBatchLogger(t)
	b, _ := NewBatcher(c, l, false)

	results := b.WriteAll(map[string]map[string]interface{}{
		"secret/x": {"foo": "bar"},
	})
	if len(results) != 1 || !results[0].Success {
		t.Errorf("expected successful write, got %+v", results)
	}
}
