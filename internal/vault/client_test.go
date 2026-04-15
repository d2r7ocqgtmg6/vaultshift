package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vaultshift/internal/vault"
)

func newMockVaultServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestNew_MissingAddress(t *testing.T) {
	_, err := vault.New(vault.Config{Token: "tok"})
	if err == nil {
		t.Fatal("expected error for missing address")
	}
}

func TestNew_MissingToken(t *testing.T) {
	_, err := vault.New(vault.Config{Address: "http://localhost:8200"})
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestNew_Success(t *testing.T) {
	c, err := vault.New(vault.Config{
		Address:   "http://localhost:8200",
		Token:     "root",
		Namespace: "ns1",
		Timeout:   5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Namespace() != "ns1" {
		t.Errorf("expected namespace ns1, got %s", c.Namespace())
	}
}

func TestReadSecret_NotFound(t *testing.T) {
	srv := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	c, _ := vault.New(vault.Config{Address: srv.URL, Token: "root"})
	_, err := c.ReadSecret("secret/data/missing")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestListSecrets_Empty(t *testing.T) {
	srv := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{},
		})
	})
	defer srv.Close()

	c, _ := vault.New(vault.Config{Address: srv.URL, Token: "root"})
	keys, err := c.ListSecrets("secret/metadata/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}
