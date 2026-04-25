package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newGraphMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/v1/secret/metadata/root":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"child/", "leaf"}},
			})
		case "/v1/secret/metadata/root/child/":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"secret"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newGraphClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewGrapher_MissingClient(t *testing.T) {
	_, err := NewGrapher(nil, newTraceLogger(t))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewGrapher_MissingLogger(t *testing.T) {
	srv := newGraphMockServer(t)
	defer srv.Close()
	c := newGraphClient(t, srv)
	_, err := NewGrapher(c, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGraph_Build_ReturnsNodes(t *testing.T) {
	srv := newGraphMockServer(t)
	defer srv.Close()
	c := newGraphClient(t, srv)
	g, err := NewGrapher(c, newTraceLogger(t))
	if err != nil {
		t.Fatalf("NewGrapher: %v", err)
	}
	result, err := g.Build("secret/metadata/root")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(result.Nodes) == 0 {
		t.Fatal("expected at least one node")
	}
	if len(result.Edges) == 0 {
		t.Fatal("expected at least one edge")
	}
}

func TestGraph_Build_EdgeSources(t *testing.T) {
	srv := newGraphMockServer(t)
	defer srv.Close()
	c := newGraphClient(t, srv)
	g, _ := NewGrapher(c, newTraceLogger(t))
	result, err := g.Build("secret/metadata/root")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, edge := range result.Edges {
		if strings.TrimSpace(edge[0]) == "" || strings.TrimSpace(edge[1]) == "" {
			t.Errorf("edge has empty endpoint: %v", edge)
		}
	}
}
