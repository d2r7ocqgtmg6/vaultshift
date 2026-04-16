package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSearchMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/ns":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db"}},
			})
		case "/v1/secret/data/ns/db":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"password": "s3cr3t", "user": "admin"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newSearchClient(addr string) *Client {
	c, _ := New(addr, "token")
	return c
}

func TestNewSearcher_MissingClient(t *testing.T) {
	_, err := NewSearcher(nil)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestSearch_MatchesKey(t *testing.T) {
	srv := newSearchMockServer()
	defer srv.Close()
	c := newSearchClient(srv.URL)
	s, _ := NewSearcher(c)

	results, err := s.Search("secret/metadata/ns", "password", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "password" {
		t.Errorf("expected key 'password', got %s", results[0].Key)
	}
}

func TestSearch_MatchesValue(t *testing.T) {
	srv := newSearchMockServer()
	defer srv.Close()
	c := newSearchClient(srv.URL)
	s, _ := NewSearcher(c)

	results, err := s.Search("secret/metadata/ns", "s3cr3t", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestSearch_NoMatch(t *testing.T) {
	srv := newSearchMockServer()
	defer srv.Close()
	c := newSearchClient(srv.URL)
	s, _ := NewSearcher(c)

	results, err := s.Search("secret/metadata/ns", "nomatch", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
