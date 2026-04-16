package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/drew/vaultshift/internal/audit"
)

func newExportMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/exports":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"keyA"}},
			})
		case "/v1/secret/data/exports/keyA":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"val": "1"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newExportLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, _ := audit.New("")
	return l
}

func newExportClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "tok", "secret")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewExporter_MissingClient(t *testing.T) {
	_, err := NewExporter(nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExport_DryRun_NoFile(t *testing.T) {
	srv := newExportMockServer()
	defer srv.Close()
	c := newExportClient(t, srv)
	ex, _ := NewExporter(c, newExportLogger(t))

	n, err := ex.Export("secret/exports", "/tmp/should-not-exist.json", true)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 got %d", n)
	}
	if _, err := os.Stat("/tmp/should-not-exist.json"); !os.IsNotExist(err) {
		os.Remove("/tmp/should-not-exist.json")
		t.Fatal("file should not exist in dry-run")
	}
}

func TestExport_WritesFile(t *testing.T) {
	srv := newExportMockServer()
	defer srv.Close()
	c := newExportClient(t, srv)
	ex, _ := NewExporter(c, newExportLogger(t))

	dest := filepath.Join(t.TempDir(), "out.json")
	n, err := ex.Export("secret/exports", dest, false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 got %d", n)
	}
	f, err := os.Open(dest)
	if err != nil {
		t.Fatal("file not created:", err)
	}
	defer f.Close()
	var result map[string]map[string]interface{}
	if err := json.NewDecoder(f).Decode(&result); err != nil {
		t.Fatal("invalid json:", err)
	}
	if _, ok := result["secret/exports/keyA"]; !ok {
		t.Fatal("missing key in output")
	}
}
