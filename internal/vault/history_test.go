package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newHistoryMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/app/db":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{"created_time": "2024-01-01T00:00:00Z", "deletion_time": ""},
						"2": map[string]interface{}{"created_time": "2024-02-01T00:00:00Z", "deletion_time": "2024-03-01T00:00:00Z"},
					},
				},
			})
		case "/v1/secret/data/app/db":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"password": "s3cr3t"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newHistoryClient(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := New(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestNewHistorian_MissingClient(t *testing.T) {
	_, err := NewHistorian(nil, "secret")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewHistorian_MissingMount(t *testing.T) {
	srv := newHistoryMockServer(t)
	defer srv.Close()
	c := newHistoryClient(t, srv.URL)
	_, err := NewHistorian(c, "")
	if err == nil {
		t.Fatal("expected error for empty mount")
	}
}

func TestListVersions_ReturnsEntries(t *testing.T) {
	srv := newHistoryMockServer(t)
	defer srv.Close()
	c := newHistoryClient(t, srv.URL)

	h, err := NewHistorian(c, "secret")
	if err != nil {
		t.Fatalf("NewHistorian: %v", err)
	}

	entries, err := h.ListVersions(context.Background(), "app/db")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestListVersions_DeletedFlagged(t *testing.T) {
	srv := newHistoryMockServer(t)
	defer srv.Close()
	c := newHistoryClient(t, srv.URL)

	h, _ := NewHistorian(c, "secret")
	entries, err := h.ListVersions(context.Background(), "app/db")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	deleted := 0
	for _, e := range entries {
		if e.Deleted {
			deleted++
		}
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted entry, got %d", deleted)
	}
}

func TestReadVersion_ReturnsData(t *testing.T) {
	srv := newHistoryMockServer(t)
	defer srv.Close()
	c := newHistoryClient(t, srv.URL)

	h, _ := NewHistorian(c, "secret")
	entry, err := h.ReadVersion(context.Background(), "app/db", 1)
	if err != nil {
		t.Fatalf("ReadVersion: %v", err)
	}
	if entry.Data["password"] != "s3cr3t" {
		t.Fatalf("unexpected data: %v", entry.Data)
	}
}
