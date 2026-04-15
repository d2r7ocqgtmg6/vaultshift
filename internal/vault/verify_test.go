package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newVerifyMockServer(t *testing.T, routes map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, ok := routes[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func newVerifyClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestVerify_AllMatched(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	routes := map[string]map[string]interface{}{
		"/v1/secret/data/foo": data,
	}
	src := newVerifyMockServer(t, routes)
	dst := newVerifyMockServer(t, routes)
	defer src.Close()
	defer dst.Close()

	v := NewVerifier(newVerifyClient(t, src), newVerifyClient(t, dst))
	res, err := v.Verify([]string{"secret/data/foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Matched) != 1 || res.Matched[0] != "secret/data/foo" {
		t.Errorf("expected 1 matched, got %v", res.Matched)
	}
	if len(res.Missing)+len(res.Mismatch) != 0 {
		t.Errorf("expected no missing/mismatch, got missing=%v mismatch=%v", res.Missing, res.Mismatch)
	}
}

func TestVerify_MissingInDest(t *testing.T) {
	src := newVerifyMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/bar": {"k": "v"},
	})
	dst := newVerifyMockServer(t, map[string]map[string]interface{}{})
	defer src.Close()
	defer dst.Close()

	v := NewVerifier(newVerifyClient(t, src), newVerifyClient(t, dst))
	res, err := v.Verify([]string{"secret/data/bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Missing) != 1 || res.Missing[0] != "secret/data/bar" {
		t.Errorf("expected 1 missing, got %v", res.Missing)
	}
}

func TestVerify_Mismatch(t *testing.T) {
	src := newVerifyMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/baz": {"k": "original"},
	})
	dst := newVerifyMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/baz": {"k": "changed"},
	})
	defer src.Close()
	defer dst.Close()

	v := NewVerifier(newVerifyClient(t, src), newVerifyClient(t, dst))
	res, err := v.Verify([]string{"secret/data/baz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Mismatch) != 1 || res.Mismatch[0] != "secret/data/baz" {
		t.Errorf("expected 1 mismatch, got %v", res.Mismatch)
	}
}
