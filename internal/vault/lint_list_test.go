package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLint_IntegrationWithLister(t *testing.T) {
	var requestedPaths []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.RawQuery, "list=true") || r.Method == "LIST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"alpha", "beta"},
				},
			})
			return
		}
		requestedPaths = append(requestedPaths, r.URL.Path)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"token": "secret-value"},
		})
	}))
	defer srv.Close()

	client, err := New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	logger := newLintLogger(t)
	linter, err := NewLinter(client, logger, NoEmptyKeys)
	if err != nil {
		t.Fatalf("NewLinter: %v", err)
	}

	lister := NewLister(client)
	paths, err := lister.List("secret/data/app")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	var allViolations []string
	for _, p := range paths {
		res, err := linter.Lint(p)
		if err != nil {
			t.Fatalf("Lint(%q): %v", p, err)
		}
		allViolations = append(allViolations, res.Violations...)
	}

	if len(allViolations) != 0 {
		t.Fatalf("expected no violations, got %v", allViolations)
	}
	if len(requestedPaths) != len(paths) {
		t.Fatalf("expected %d reads, got %d", len(paths), len(requestedPaths))
	}
}
