package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newIndexMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/app":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"db", "api"},
				},
			})
		case "/v1/secret/data/app/db":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"password": "s3cr3t", "user": "admin"},
				},
			})
		case "/v1/secret/data/app/api":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"token": "abc123"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newIndexClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewIndexer_MissingClient(t *testing.T) {
	_, err := NewIndexer(nil)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestIndexer_Build_ReturnsEntries(t *testing.T) {
	srv := newIndexMockServer(t)
	defer srv.Close()
	c := newIndexClient(t, srv)

	idxr, err := NewIndexer(c)
	if err != nil {
		t.Fatalf("NewIndexer: %v", err)
	}

	idx, err := idxr.Build("app")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(idx) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(idx))
	}
	entry, ok := idx["secret/data/app/db"]
	if !ok {
		t.Fatal("expected entry for app/db")
	}
	if len(entry.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(entry.Keys))
	}
}

func TestIndex_Lookup_FindsKey(t *testing.T) {
	idx := Index{
		"secret/data/app/db":  {Path: "secret/data/app/db", Keys: []string{"password", "user"}},
		"secret/data/app/api": {Path: "secret/data/app/api", Keys: []string{"token"}},
	}
	results := idx.Lookup("password")
	if len(results) != 1 || results[0] != "secret/data/app/db" {
		t.Errorf("unexpected lookup results: %v", results)
	}
}

func TestIndex_Lookup_CaseInsensitive(t *testing.T) {
	idx := Index{
		"secret/data/app/db": {Path: "secret/data/app/db", Keys: []string{"Password"}},
	}
	results := idx.Lookup("password")
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestIndex_Lookup_NotFound(t *testing.T) {
	idx := Index{
		"secret/data/app/db": {Path: "secret/data/app/db", Keys: []string{"user"}},
	}
	results := idx.Lookup("token")
	if len(results) != 0 {
		t.Errorf("expected no results, got %v", results)
	}
}
