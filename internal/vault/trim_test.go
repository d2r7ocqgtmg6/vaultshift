package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newTrimLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newTrimMockServer(t *testing.T, initial map[string]interface{}, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": initial}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				*written = d
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestNewTrimmer_MissingClient(t *testing.T) {
	l := newTrimLogger(t)
	_, err := NewTrimmer(nil, l, false, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewTrimmer_MissingLogger(t *testing.T) {
	c := &Client{}
	_, err := NewTrimmer(c, nil, false, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTrim_DryRun_NoWrite(t *testing.T) {
	var written map[string]interface{}
	srv := newTrimMockServer(t, map[string]interface{}{"key": "  hello  "}, &written)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "tok"})
	l := newTrimLogger(t)
	tr, _ := NewTrimmer(c, l, true, nil)
	modified, err := tr.Trim("secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modified {
		t.Fatal("expected modified=true")
	}
	if written != nil {
		t.Fatal("expected no write in dry-run")
	}
}

func TestTrim_TrimsValues(t *testing.T) {
	var written map[string]interface{}
	srv := newTrimMockServer(t, map[string]interface{}{"key": "  hello  ", "other": "clean"}, &written)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "tok"})
	l := newTrimLogger(t)
	tr, _ := NewTrimmer(c, l, false, nil)
	modified, err := tr.Trim("secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modified {
		t.Fatal("expected modified")
	}
	if written["key"] != "hello" {
		t.Fatalf("expected trimmed value, got %q", written["key"])
	}
}

func TestTrim_NoChange(t *testing.T) {
	var written map[string]interface{}
	srv := newTrimMockServer(t, map[string]interface{}{"key": "clean"}, &written)
	defer srv.Close()
	c, _ := New(Config{Address: srv.URL, Token: "tok"})
	l := newTrimLogger(t)
	tr, _ := NewTrimmer(c, l, false, nil)
	modified, err := tr.Trim("secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if modified {
		t.Fatal("expected no modification")
	}
	if written != nil {
		t.Fatal("expected no write")
	}
}
