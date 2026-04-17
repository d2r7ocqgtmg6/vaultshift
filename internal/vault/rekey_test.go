package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newRekeyLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newRekeyMockServer(t *testing.T, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/secret/"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"old_key": "value1", "keep": "value2"},
			})
		case r.Method == http.MethodPost || r.Method == http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			*written = body
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestNewRekeyer_MissingClient(t *testing.T) {
	l := newRekeyLogger(t)
	_, err := NewRekeyer(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestRekey_DryRun_NoWrite(t *testing.T) {
	var written map[string]interface{}
	srv := newRekeyMockServer(t, &written)
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	l := newRekeyLogger(t)
	rk, _ := NewRekeyer(client, l, true)

	res, err := rk.Rekey("secret/data/mypath", map[string]string{"old_key": "new_key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written != nil {
		t.Error("expected no write in dry-run mode")
	}
	if res.Renamed["old_key"] != "new_key" {
		t.Errorf("expected old_key->new_key, got %v", res.Renamed)
	}
}

func TestRekey_WritesRenamedKeys(t *testing.T) {
	var written map[string]interface{}
	srv := newRekeyMockServer(t, &written)
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	l := newRekeyLogger(t)
	rk, _ := NewRekeyer(client, l, false)

	res, err := rk.Rekey("secret/data/mypath", map[string]string{"old_key": "new_key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Renamed) != 1 {
		t.Errorf("expected 1 renamed key, got %d", len(res.Renamed))
	}
	if len(res.Skipped) != 1 || res.Skipped[0] != "keep" {
		t.Errorf("expected 'keep' in skipped, got %v", res.Skipped)
	}
}
