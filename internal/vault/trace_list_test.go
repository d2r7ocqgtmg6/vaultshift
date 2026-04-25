package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newTraceListMockServer(t *testing.T, paths []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		for _, p := range paths {
			if r.URL.Path == "/v1/"+p {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"k": "v"},
				})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestTrace_MultiplePathsAllRecorded(t *testing.T) {
	paths := []string{"secret/data/a", "secret/data/b", "secret/data/c"}
	srv := newTraceListMockServer(t, paths)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l, _ := audit.New("")
	tracer, _ := NewTracer(client, l)

	for _, p := range paths {
		tracer.Trace(p)
	}

	if len(tracer.Entries()) != len(paths) {
		t.Errorf("expected %d entries, got %d", len(paths), len(tracer.Entries()))
	}
}

func TestTrace_MixedResults(t *testing.T) {
	srv := newTraceListMockServer(t, []string{"secret/data/exists"})
	defer srv.Close()

	client, _ := New(srv.URL, "token")
	l, _ := audit.New("")
	tracer, _ := NewTracer(client, l)

	ok := tracer.Trace("secret/data/exists")
	miss := tracer.Trace("secret/data/missing")

	if ok.Error != "" {
		t.Errorf("expected no error for existing path, got: %s", ok.Error)
	}
	if miss.Error == "" {
		t.Error("expected error for missing path")
	}
	if len(tracer.Entries()) != 2 {
		t.Errorf("expected 2 entries, got %d", len(tracer.Entries()))
	}
}
