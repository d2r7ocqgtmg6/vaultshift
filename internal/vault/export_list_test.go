package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExport_IntegrationWithLister(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/multi":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"a", "b"}},
			})
		case "/v1/secret/data/multi/a", "/v1/secret/data/multi/b":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"x": "y"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := New(srv.URL, "tok", "secret")
	if err != nil {
		t.Fatal(err)
	}
	ex, err := NewExporter(c, nil)
	if err != nil {
		t.Fatal(err)
	}

	n, err := ex.Export("secret/multi", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 secrets, got %d", n)
	}
}
