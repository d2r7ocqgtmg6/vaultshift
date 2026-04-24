package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func newPrefetchMockServer(secrets map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if val, ok := secrets[path]; ok {
			fmt.Fprintf(w, `{"data":{"data":{"value":%q}}}`, val)
			return
		}
		http.Error(w, `{"errors":[]}`, http.StatusNotFound)
	}))
}

func newPrefetchClient(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := New(addr, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestNewPrefetcher_MissingClient(t *testing.T) {
	_, err := NewPrefetcher(nil, 2)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewPrefetcher_InvalidConcurrency(t *testing.T) {
	svr := newPrefetchMockServer(nil)
	defer svr.Close()
	c := newPrefetchClient(t, svr.URL)

	_, err := NewPrefetcher(c, 0)
	if err == nil {
		t.Fatal("expected error for concurrency < 1")
	}
}

func TestPrefetch_ReturnsAllResults(t *testing.T) {
	svr := newPrefetchMockServer(map[string]string{
		"/v1/secret/data/a": "alpha",
		"/v1/secret/data/b": "beta",
	})
	defer svr.Close()
	c := newPrefetchClient(t, svr.URL)

	p, err := NewPrefetcher(c, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := p.Fetch([]string{"secret/data/a", "secret/data/b", "secret/data/missing"})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	var errCount int
	for _, r := range results {
		if r.Err != nil {
			errCount++
		}
	}
	if errCount != 1 {
		t.Errorf("expected 1 error result, got %d", errCount)
	}
}

func TestPrefetch_FetchMap_SeparatesErrors(t *testing.T) {
	svr := newPrefetchMockServer(map[string]string{
		"/v1/secret/data/x": "xval",
	})
	defer svr.Close()
	c := newPrefetchClient(t, svr.URL)

	p, _ := NewPrefetcher(c, 1)
	out, errs := p.FetchMap([]string{"secret/data/x", "secret/data/y"})

	if len(out) != 1 {
		t.Errorf("expected 1 success, got %d", len(out))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestPrefetch_OrderPreserved(t *testing.T) {
	paths := []string{"secret/data/c", "secret/data/a", "secret/data/b"}
	svr := newPrefetchMockServer(map[string]string{
		"/v1/secret/data/a": "a",
		"/v1/secret/data/b": "b",
		"/v1/secret/data/c": "c",
	})
	defer svr.Close()
	c := newPrefetchClient(t, svr.URL)

	p, _ := NewPrefetcher(c, 3)
	results := p.Fetch(paths)

	got := make([]string, len(results))
	for i, r := range results {
		got[i] = r.Path
	}
	expected := append([]string(nil), paths...)
	if !sort.StringsAreSorted(got) || fmt.Sprint(got) != fmt.Sprint(expected) {
		// order must match input regardless of goroutine scheduling
		if fmt.Sprint(got) != fmt.Sprint(expected) {
			t.Errorf("order mismatch: got %v, want %v", got, expected)
		}
	}
}
