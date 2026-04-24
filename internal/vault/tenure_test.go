package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vaultshift/internal/audit"
)

func newTenureLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newTenureMockServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		data, ok := secrets[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func TestNewTenurer_MissingClient(t *testing.T) {
	l := newTenureLogger(t)
	_, err := NewTenurer(nil, l, 30)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTenurer_MissingLogger(t *testing.T) {
	srv := newTenureMockServer(t, nil)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	_, err := NewTenurer(c, nil, 30)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewTenurer_InvalidMaxDays(t *testing.T) {
	srv := newTenureMockServer(t, nil)
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTenureLogger(t)
	_, err := NewTenurer(c, l, 0)
	if err == nil {
		t.Fatal("expected error for maxDays=0")
	}
}

func TestTenure_BelowMax(t *testing.T) {
	recent := time.Now().UTC().Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	srv := newTenureMockServer(t, map[string]map[string]interface{}{
		"secret/foo": {"created_at": recent, "value": "bar"},
	})
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTenureLogger(t)
	tn, _ := NewTenurer(c, l, 30)

	results, err := tn.Check([]string{"secret/foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ExceedsMax {
		t.Error("expected ExceedsMax=false for a 10-day-old secret with max=30")
	}
}

func TestTenure_ExceedsMax(t *testing.T) {
	old := time.Now().UTC().Add(-60 * 24 * time.Hour).Format(time.RFC3339)
	srv := newTenureMockServer(t, map[string]map[string]interface{}{
		"secret/old": {"created_at": old, "value": "stale"},
	})
	defer srv.Close()
	c, _ := New(srv.URL, "tok")
	l := newTenureLogger(t)
	tn, _ := NewTenurer(c, l, 30)

	results, err := tn.Check([]string{"secret/old"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results[0].ExceedsMax {
		t.Error("expected ExceedsMax=true for a 60-day-old secret with max=30")
	}
}
