package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wearevault/vaultshift/internal/audit"
)

func newDigestMockServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for k, v := range secrets {
			if path == "/v1/"+k {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": v})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newDigestClient(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := New(addr, "token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func newDigestLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewDigester_MissingClient(t *testing.T) {
	l := newDigestLogger(t)
	_, err := NewDigester(nil, l)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewDigester_MissingLogger(t *testing.T) {
	srv := newDigestMockServer(t, nil)
	defer srv.Close()
	c := newDigestClient(t, srv.URL)
	_, err := NewDigester(c, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestDigest_ReturnsHex(t *testing.T) {
	srv := newDigestMockServer(t, map[string]map[string]interface{}{
		"secret/foo": {"key": "value"},
	})
	defer srv.Close()
	c := newDigestClient(t, srv.URL)
	l := newDigestLogger(t)
	d, err := NewDigester(c, l)
	if err != nil {
		t.Fatalf("NewDigester: %v", err)
	}
	res, err := d.Digest("secret/foo")
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	if len(res.Digest) != 64 {
		t.Errorf("expected 64-char hex digest, got %d chars", len(res.Digest))
	}
}

func TestDigest_Deterministic(t *testing.T) {
	data := map[string]interface{}{"b": "2", "a": "1"}
	h1, _ := computeDigest(data)
	h2, _ := computeDigest(data)
	if h1 != h2 {
		t.Errorf("digest not deterministic: %s != %s", h1, h2)
	}
}

func TestDigest_NotFound(t *testing.T) {
	srv := newDigestMockServer(t, nil)
	defer srv.Close()
	c := newDigestClient(t, srv.URL)
	l := newDigestLogger(t)
	d, _ := NewDigester(c, l)
	_, err := d.Digest("secret/missing")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}
