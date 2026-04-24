package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newNormalizeMockServer(t *testing.T, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			payload := map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"API_KEY": "  secret123  ",
						"HOST":    "  localhost  ",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payload)
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if written != nil {
				if d, ok := body["data"].(map[string]interface{}); ok {
					*written = d
				}
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func newNormalizeLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewNormalizer_MissingClient(t *testing.T) {
	_, err := NewNormalizer(nil, newNormalizeLogger(t))
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewNormalizer_MissingLogger(t *testing.T) {
	svr := newNormalizeMockServer(t, nil)
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	_, err := NewNormalizer(c, nil)
	if err == nil || !strings.Contains(err.Error(), "logger") {
		t.Fatalf("expected logger error, got %v", err)
	}
}

func TestNormalize_DryRun_NoWrite(t *testing.T) {
	var written map[string]interface{}
	svr := newNormalizeMockServer(t, &written)
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	n, _ := NewNormalizer(c, newNormalizeLogger(t), WithNormalizeDryRun(), WithNormalizeTrimSpace())
	if err := n.Normalize("secret/data/app"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written != nil {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestNormalize_TrimsAndLowers(t *testing.T) {
	var written map[string]interface{}
	svr := newNormalizeMockServer(t, &written)
	defer svr.Close()
	c, _ := New(svr.URL, "tok")
	n, _ := NewNormalizer(c, newNormalizeLogger(t),
		WithNormalizeTrimSpace(),
		WithNormalizeLowerKeys(),
	)
	if err := n.Normalize("secret/data/app"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written == nil {
		t.Fatal("expected write to have occurred")
	}
	if v, ok := written["api_key"]; !ok || v != "secret123" {
		t.Fatalf("expected api_key=secret123, got %v", written)
	}
	if v, ok := written["host"]; !ok || v != "localhost" {
		t.Fatalf("expected host=localhost, got %v", written)
	}
}
