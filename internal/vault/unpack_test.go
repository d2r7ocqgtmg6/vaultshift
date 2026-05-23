package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newUnpackLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newUnpackMockServer(t *testing.T, written *map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	*written = make(map[string]map[string]interface{})
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]interface{})
			(*written)[r.URL.Path] = data
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestNewUnpacker_MissingClient(t *testing.T) {
	l := newUnpackLogger(t)
	_, err := NewUnpacker(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewUnpacker_MissingLogger(t *testing.T) {
	c := &Client{}
	_, err := NewUnpacker(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestUnpack_DryRun_NoWrite(t *testing.T) {
	var written map[string]map[string]interface{}
	srv := newUnpackMockServer(t, &written)
	defer srv.Close()

	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l := newUnpackLogger(t)
	u, err := NewUnpacker(c, l, true)
	if err != nil {
		t.Fatalf("NewUnpacker: %v", err)
	}

	payload := map[string]map[string]interface{}{
		"secret/app/db": {"password": "s3cr3t"},
	}
	raw, _ := json.Marshal(payload)
	f := filepath.Join(t.TempDir(), "pack.json")
	_ = os.WriteFile(f, raw, 0600)

	if err := u.Unpack(f); err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if len(written) != 0 {
		t.Errorf("expected no writes in dry-run, got %d", len(written))
	}
}

func TestUnpack_WritesSecrets(t *testing.T) {
	var written map[string]map[string]interface{}
	srv := newUnpackMockServer(t, &written)
	defer srv.Close()

	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l := newUnpackLogger(t)
	u, err := NewUnpacker(c, l, false)
	if err != nil {
		t.Fatalf("NewUnpacker: %v", err)
	}

	payload := map[string]map[string]interface{}{
		"secret/app/db": {"password": "s3cr3t"},
	}
	raw, _ := json.Marshal(payload)
	f := filepath.Join(t.TempDir(), "pack.json")
	_ = os.WriteFile(f, raw, 0600)

	if err := u.Unpack(f); err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if len(written) != 1 {
		t.Errorf("expected 1 write, got %d", len(written))
	}
}
