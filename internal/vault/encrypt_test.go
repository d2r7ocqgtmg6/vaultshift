package vault

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newEncryptLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newEncryptMockServer(t *testing.T, data map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"data": data},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestNewEncrypter_MissingClient(t *testing.T) {
	l := newEncryptLogger(t)
	_, err := NewEncrypter(nil, l, make([]byte, 32), false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewEncrypter_BadKeyLength(t *testing.T) {
	l := newEncryptLogger(t)
	srv := newEncryptMockServer(t, nil)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewEncrypter(c, l, make([]byte, 10), false)
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestEncrypt_DryRun_NoWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"data": map[string]any{"password": "secret"}},
		})
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "tok")
	l := newEncryptLogger(t)
	enc, err := NewEncrypter(c, l, make([]byte, 32), true)
	if err != nil {
		t.Fatalf("NewEncrypter: %v", err)
	}
	if err := enc.Encrypt("secret/my-app"); err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if writes != 0 {
		t.Errorf("expected 0 writes in dry-run, got %d", writes)
	}
}

func TestEncrypt_WritesEncryptedValues(t *testing.T) {
	var written map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				written, _ = d.(map[string]any)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"data": map[string]any{"api_key": "hunter2"}},
		})
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "tok")
	l := newEncryptLogger(t)
	key := make([]byte, 32)
	enc, _ := NewEncrypter(c, l, key, false)
	if err := enc.Encrypt("secret/svc"); err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if written == nil {
		t.Fatal("expected a write to have occurred")
	}
	val, ok := written["api_key"].(string)
	if !ok {
		t.Fatalf("expected string value, got %T", written["api_key"])
	}
	if _, err := base64.StdEncoding.DecodeString(val); err != nil {
		t.Errorf("written value is not valid base64: %v", err)
	}
}
