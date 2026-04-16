package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newImportLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newImportMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewImporter_MissingClient(t *testing.T) {
	l := newImportLogger(t)
	_, err := NewImporter(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestImport_DryRun_NoWrite(t *testing.T) {
	srv := newImportMockServer(t)
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	logger := newImportLogger(t)
	imp, _ := NewImporter(client, logger, true)

	entries := []ImportEntry{
		{Path: "secret/foo", Data: map[string]interface{}{"key": "val"}},
	}
	f, _ := os.CreateTemp(t.TempDir(), "import*.json")
	json.NewEncoder(f).Encode(entries)
	f.Close()

	n, err := imp.Import(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 imported, got %d", n)
	}
}

func TestImport_WritesSecrets(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	logger := newImportLogger(t)
	imp, _ := NewImporter(client, logger, false)

	entries := []ImportEntry{
		{Path: "secret/a", Data: map[string]interface{}{"x": "1"}},
		{Path: "secret/b", Data: map[string]interface{}{"y": "2"}},
	}
	f, _ := os.CreateTemp(t.TempDir(), "import*.json")
	json.NewEncoder(f).Encode(entries)
	f.Close()

	n, err := imp.Import(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2, got %d", n)
	}
}

func TestImport_InvalidFile(t *testing.T) {
	client, _ := New("http://127.0.0.1", "token")
	logger := newImportLogger(t)
	imp, _ := NewImporter(client, logger, false)
	_, err := imp.Import(context.Background(), "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
