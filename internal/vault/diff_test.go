package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newDiffMockServer(srcData, dstData map[string]interface{}) (*httptest.Server, *httptest.Server) {
	makeServer := func(data map[string]interface{}) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Path
			val, ok := data[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": val})
		}))
	}
	return makeServer(srcData), makeServer(dstData)
}

func newDiffClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	c, err := New(server.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestDiff_NewSecret(t *testing.T) {
	srcSrv, dstSrv := newDiffMockServer(
		map[string]interface{}{"/v1/secret/data/foo": map[string]string{"value": "bar"}},
		map[string]interface{}{},
	)
	defer srcSrv.Close()
	defer dstSrv.Close()

	d := NewDiffer(newDiffClient(t, srcSrv), newDiffClient(t, dstSrv))
	result, err := d.Diff("secret/data/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != DiffStatusNew {
		t.Errorf("expected status %q, got %q", DiffStatusNew, result.Status)
	}
}

func TestDiff_OrphanedSecret(t *testing.T) {
	srcSrv, dstSrv := newDiffMockServer(
		map[string]interface{}{},
		map[string]interface{}{"/v1/secret/data/foo": map[string]string{"value": "bar"}},
	)
	defer srcSrv.Close()
	defer dstSrv.Close()

	d := NewDiffer(newDiffClient(t, srcSrv), newDiffClient(t, dstSrv))
	result, err := d.Diff("secret/data/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != DiffStatusOrphaned {
		t.Errorf("expected status %q, got %q", DiffStatusOrphaned, result.Status)
	}
}

func TestDiff_BothMissing_ReturnsError(t *testing.T) {
	srcSrv, dstSrv := newDiffMockServer(
		map[string]interface{}{},
		map[string]interface{}{},
	)
	defer srcSrv.Close()
	defer dstSrv.Close()

	d := NewDiffer(newDiffClient(t, srcSrv), newDiffClient(t, dstSrv))
	_, err := d.Diff("secret/data/missing")
	if err == nil {
		t.Error("expected error when secret missing in both, got nil")
	}
}
