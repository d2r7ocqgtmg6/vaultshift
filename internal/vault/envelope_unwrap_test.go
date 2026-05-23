package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadEnvelope_InvalidFile(t *testing.T) {
	_, err := LoadEnvelope("/nonexistent/path/envelope.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadEnvelope_RoundTrip(t *testing.T) {
	srv := newEnvelopeMockServer(t)
	defer srv.Close()
	c := newEnvelopeClient(t, srv)
	l := newEnvelopeLogger(t)

	e, _ := NewEnveloper(c, l, false)
	out := filepath.Join(t.TempDir(), "rt.json")
	original, err := e.Wrap([]string{"secret/data/foo"}, out)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}

	loaded, err := LoadEnvelope(out)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Entries[0].Path != original.Entries[0].Path {
		t.Errorf("path mismatch: got %s want %s", loaded.Entries[0].Path, original.Entries[0].Path)
	}
}

func TestEnvelope_EntryTimestamp(t *testing.T) {
	srv := newEnvelopeMockServer(t)
	defer srv.Close()
	c := newEnvelopeClient(t, srv)
	l := newEnvelopeLogger(t)

	before := time.Now().UTC().Add(-time.Second)
	e, _ := NewEnveloper(c, l, true)
	env, err := e.Wrap([]string{"secret/data/foo"}, "")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	after := time.Now().UTC().Add(time.Second)

	ts := env.Entries[0].CreatedAt
	if ts.Before(before) || ts.After(after) {
		t.Errorf("unexpected timestamp: %v", ts)
	}
}

func TestEnvelope_ReadError_Propagates(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "tok")
	l := newEnvelopeLogger(t)
	e, _ := NewEnveloper(c, l, false)

	_, err := e.Wrap([]string{"secret/data/missing"}, "/tmp/noop.json")
	if err == nil {
		t.Fatal("expected error on read failure")
	}
}

func TestEnvelope_MultipleEntries_OrderPreserved(t *testing.T) {
	paths := []string{"secret/data/foo", "secret/data/bar"}
	counter := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"n": counter}},
		})
	}))
	defer srv.Close()

	c, _ := New(srv.URL, "tok")
	l := newEnvelopeLogger(t)
	e, _ := NewEnveloper(c, l, true)

	env, err := e.Wrap(paths, "")
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	for i, p := range paths {
		if env.Entries[i].Path != p {
			t.Errorf("entry %d: got path %s want %s", i, env.Entries[i].Path, p)
		}
	}
}
