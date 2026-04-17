package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newCompareMockServer(t *testing.T, routes map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, ok := routes[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func newCompareClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestNewComparer_MissingClient(t *testing.T) {
	_, err := NewComparer(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil src client")
	}
}

func TestCompare_Match(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	srv := newCompareMockServer(t, map[string]map[string]interface{}{"/secret/foo": data})
	defer srv.Close()

	src := newCompareClient(t, srv.URL)
	dst := newCompareClient(t, srv.URL)

	cmp, err := NewComparer(src, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, err := cmp.Compare("secret/foo")
	if err != nil {
		t.Fatalf("compare error: %v", err)
	}
	if !res.Match {
		t.Error("expected secrets to match")
	}
}

func TestCompare_MissingInDst(t *testing.T) {
	srv := newCompareMockServer(t, map[string]map[string]interface{}{"/secret/foo": {"k": "v"}})
	defer srv.Close()
	emptySrv := newCompareMockServer(t, map[string]map[string]interface{}{})
	defer emptySrv.Close()

	src := newCompareClient(t, srv.URL)
	dst := newCompareClient(t, emptySrv.URL)

	cmp, _ := NewComparer(src, dst)
	res, err := cmp.Compare("secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Missing != "dst" {
		t.Errorf("expected missing=dst, got %q", res.Missing)
	}
}
