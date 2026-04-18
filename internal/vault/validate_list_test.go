package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidate_IntegrationWithLister(t *testing.T) {
	paths := []string{"secret/a", "secret/b"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list") == "true" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"a", "b"}},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"token": "abc123"},
		})
	}))
	defer srv.Close()

	c, err := New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l := newValidateLogger(t)
	v, err := NewValidator(c, l, []ValidationRule{RequiredKeys("token")}, true)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	for _, p := range paths {
		results, err := v.Validate(p)
		if err != nil {
			t.Fatalf("Validate %s: %v", p, err)
		}
		if len(results) == 0 || !results[0].Passed {
			t.Fatalf("expected pass for %s", p)
		}
	}
}
