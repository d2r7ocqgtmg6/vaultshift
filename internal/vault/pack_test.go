package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newPackMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"key": "value"},
		})
	}))
}

func newPackClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	c.SetToken("test-token")
	return c
}

func newPackLogger(t *testing.T) AuditLogger {
	t.Helper()
	l, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestNewPacker_MissingClient(t *testing.T) {
	_, err := NewPacker(nil, newPackLogger(t), false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewPacker_MissingLogger(t *testing.T) {
	svr := newPackMockServer(t)
	defer svr.Close()
	c := newPackClient(t, svr.URL)
	_, err := NewPacker(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestPack_DryRun_NoFile(t *testing.T) {
	svr := newPackMockServer(t)
	defer svr.Close()
	c := newPackClient(t, svr.URL)
	p, err := NewPacker(c, newPackLogger(t), true)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/pack.json"
	if err := p.Pack([]string{"secret/data/foo"}, out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Error("expected no file written in dry-run mode")
	}
}

func TestPack_WritesFile(t *testing.T) {
	svr := newPackMockServer(t)
	defer svr.Close()
	c := newPackClient(t, svr.URL)
	p, err := NewPacker(c, newPackLogger(t), false)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/pack.json"
	if err := p.Pack([]string{"secret/data/foo", "secret/data/bar"}, out); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var entries []PackEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}
