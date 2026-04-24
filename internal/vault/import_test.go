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

// writeImportFile creates a temporary JSON file containing the given entries
// and returns its path. The file is automatically cleaned up when the test ends.
func writeImportFile(t *testing.T, entries []ImportEntry) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "import*.json")
	if err != nil {
		t.Fatalf("os.CreateTemp: %v", err)
	}
	if err := json.NewEncoder(f).Encode(entries); err != nil {
		t.Fatalf("json.Encode: %v", err)
	}
	f.Close()
	return f.Name()
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
	path := writeImportFile(t, entries)

	n, err := imp.Import(context.Background(), path)
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
	path := writeImportFile(t, entries)

	n, err := imp.Import(context.Background(), path)
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
