package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMigrate_RecordsRollbackOnSuccess verifies that a successful write is
// registered with the Rollbacker.
func TestMigrate_RecordsRollbackOnSuccess(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{"key": "val"},
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || strings.Contains(r.URL.Path, "/data/") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": payload})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer svr.Close()

	src, _ := New(svr.URL, "src-token")
	dst, _ := New(svr.URL, "dst-token")
	l := newRollbackLogger(t)
	rb := NewRollbacker(dst, l)
	mig := NewMigrator(src, dst, l, rb)

	err := mig.Migrate(context.Background(), "secret", "foo", "secret", "foo", MigrateOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rb.Len() != 1 {
		t.Fatalf("expected 1 rollback record, got %d", rb.Len())
	}
}

// TestMigrate_DryRun_NoRollbackRecord ensures dry-run writes are not recorded.
func TestMigrate_DryRun_NoRollbackRecord(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{"key": "val"},
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": payload})
	}))
	defer svr.Close()

	src, _ := New(svr.URL, "src-token")
	dst, _ := New(svr.URL, "dst-token")
	l := newRollbackLogger(t)
	rb := NewRollbacker(dst, l)
	mig := NewMigrator(src, dst, l, rb)

	err := mig.Migrate(context.Background(), "secret", "bar", "secret", "bar", MigrateOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rb.Len() != 0 {
		t.Fatalf("expected 0 rollback records, got %d", rb.Len())
	}
}
