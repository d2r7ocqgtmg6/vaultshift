package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newEnvelopeMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/foo":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}},
			})
		case "/v1/secret/data/bar":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"token": "abc"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newEnvelopeLogger(t *testing.T) *AuditLogger {
	t.Helper()
	l, err := New("")
	if err != nil {
		t.Fatalf("audit logger: %v", err)
	}
	return l
}

func newEnvelopeClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return c
}

func TestNewEnveloper_MissingClient(t *testing.T) {
	l := newEnvelopeLogger(t)
	_, err := NewEnveloper(nil, l, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewEnveloper_MissingLogger(t *testing.T) {
	srv := newEnvelopeMockServer(t)
	defer srv.Close()
	c := newEnvelopeClient(t, srv)
	_, err := NewEnveloper(c, nil, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestEnvelope_DryRun_NoFile(t *testing.T) {
	srv := newEnvelopeMockServer(t)
	defer srv.Close()
	c := newEnvelopeClient(t, srv)
	l := newEnvelopeLogger(t)

	e, err := NewEnveloper(c, l, true)
	if err != nil {
		t.Fatalf("new enveloper: %v", err)
	}

	out := filepath.Join(t.TempDir(), "envelope.json")
	env, err := e.Wrap([]string{"secret/data/foo"}, out)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if len(env.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(env.Entries))
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("expected no file in dry-run mode")
	}
}

func TestEnvelope_WritesFile(t *testing.T) {
	srv := newEnvelopeMockServer(t)
	defer srv.Close()
	c := newEnvelopeClient(t, srv)
	l := newEnvelopeLogger(t)

	e, err := NewEnveloper(c, l, false)
	if err != nil {
		t.Fatalf("new enveloper: %v", err)
	}

	out := filepath.Join(t.TempDir(), "envelope.json")
	env, err := e.Wrap([]string{"secret/data/foo", "secret/data/bar"}, out)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if len(env.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(env.Entries))
	}

	loaded, err := LoadEnvelope(out)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Entries) != 2 {
		t.Fatalf("loaded entries mismatch: got %d", len(loaded.Entries))
	}
}
