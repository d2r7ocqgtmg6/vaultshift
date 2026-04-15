package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newRollbackLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestRollbacker_RecordAndLen(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	c, _ := New(svr.URL, "test-token")
	rb := NewRollbacker(c, newRollbackLogger(t))

	if rb.Len() != 0 {
		t.Fatalf("expected 0 records, got %d", rb.Len())
	}
	rb.Record("secret", "foo/bar")
	rb.Record("secret", "foo/baz")
	if rb.Len() != 2 {
		t.Fatalf("expected 2 records, got %d", rb.Len())
	}
}

func TestRollbacker_Rollback_Success(t *testing.T) {
	deleted := []string{}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted = append(deleted, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer svr.Close()

	c, _ := New(svr.URL, "test-token")
	rb := NewRollbacker(c, newRollbackLogger(t))
	rb.Record("secret", "a")
	rb.Record("secret", "b")

	errs := rb.Rollback(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(deleted) != 2 {
		t.Fatalf("expected 2 deletes, got %d", len(deleted))
	}
}

func TestRollbacker_Rollback_PartialError(t *testing.T) {
	calls := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer svr.Close()

	c, _ := New(svr.URL, "test-token")
	rb := NewRollbacker(c, newRollbackLogger(t))
	rb.Record("secret", "x")
	rb.Record("secret", "y")

	errs := rb.Rollback(context.Background())
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}
