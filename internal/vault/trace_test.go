package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newTraceMockServer(t *testing.T, path string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/"+path {
			w.WriteHeader(statusCode)
			if statusCode == http.StatusOK {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"key": "val"},
				})
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTraceLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewTracer_MissingClient(t *testing.T) {
	_, err := NewTracer(nil, newTraceLogger(t))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTracer_MissingLogger(t *testing.T) {
	client := &Client{}
	_, err := NewTracer(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTrace_RecordsEntry(t *testing.T) {
	srv := newTraceMockServer(t, "secret/data/foo", http.StatusOK)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tracer, err := NewTracer(client, newTraceLogger(t))
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}

	entry := tracer.Trace("secret/data/foo")
	if entry.Path != "secret/data/foo" {
		t.Errorf("expected path secret/data/foo, got %s", entry.Path)
	}
	if entry.Operation != "read" {
		t.Errorf("expected operation read, got %s", entry.Operation)
	}
	if len(tracer.Entries()) != 1 {
		t.Errorf("expected 1 entry, got %d", len(tracer.Entries()))
	}
}

func TestTrace_RecordsError(t *testing.T) {
	srv := newTraceMockServer(t, "secret/data/foo", http.StatusNotFound)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tracer, _ := NewTracer(client, newTraceLogger(t))

	entry := tracer.Trace("secret/data/missing")
	if entry.Error == "" {
		t.Error("expected error to be recorded for not-found path")
	}
}
