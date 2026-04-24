package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newStampMockServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	writes := &[]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			*writes = append(*writes, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	return srv, writes
}

func newStampClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("api.NewClient: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestNewStamper_MissingClient(t *testing.T) {
	_, err := NewStamper(nil, newAuditLogger(t))
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestNewStamper_EmptyField(t *testing.T) {
	srv, _ := newStampMockServer(t)
	defer srv.Close()
	c := newStampClient(t, srv.URL)
	_, err := NewStamper(c, newAuditLogger(t), WithStampField(""))
	if err == nil || !strings.Contains(err.Error(), "field") {
		t.Fatalf("expected field error, got %v", err)
	}
}

func TestStamp_DryRun_NoWrite(t *testing.T) {
	srv, writes := newStampMockServer(t)
	defer srv.Close()
	c := newStampClient(t, srv.URL)
	s, err := NewStamper(c, newAuditLogger(t), WithStampDryRun(true))
	if err != nil {
		t.Fatalf("NewStamper: %v", err)
	}
	if err := s.Stamp("secret/myapp/config"); err != nil {
		t.Fatalf("Stamp: %v", err)
	}
	if len(*writes) != 0 {
		t.Fatalf("expected no writes in dry-run, got %d", len(*writes))
	}
}

func TestStamp_WritesTimestamp(t *testing.T) {
	srv, writes := newStampMockServer(t)
	defer srv.Close()
	c := newStampClient(t, srv.URL)
	s, err := NewStamper(c, newAuditLogger(t), WithStampField("stamped_at"))
	if err != nil {
		t.Fatalf("NewStamper: %v", err)
	}
	if err := s.Stamp("secret/myapp/config"); err != nil {
		t.Fatalf("Stamp: %v", err)
	}
	if len(*writes) != 1 {
		t.Fatalf("expected 1 write, got %d", len(*writes))
	}
}
