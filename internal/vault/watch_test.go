package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newWatchMockServer(t *testing.T, responses []map[string]interface{}) *httptest.Server {
	t.Helper()
	var call int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(atomic.AddInt32(&call, 1)) - 1
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": responses[idx],
		})
	}))
}

func TestNewWatcher_MissingClient(t *testing.T) {
	_, err := NewWatcher(nil, nil, time.Second, []string{"secret/foo"})
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewWatcher_NoPaths(t *testing.T) {
	srv := newWatchMockServer(t, []map[string]interface{}{{"k": "v"}})
	defer srv.Close()
	c, _ := New(srv.URL, "token")
	_, err := NewWatcher(c, nil, time.Second, nil)
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

func TestWatch_DetectsChange(t *testing.T) {
	responses := []map[string]interface{}{
		{"password": "old"},
		{"password": "new"},
	}
	srv := newWatchMockServer(t, responses)
	defer srv.Close()

	c, err := New(srv.URL, "token")
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	w, err := NewWatcher(c, nil, 20*time.Millisecond, []string{"secret/foo"})
	if err != nil {
		t.Fatalf("watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ch := w.Watch(ctx)
	var events []WatchEvent
	for e := range ch {
		events = append(events, e)
		if len(events) >= 2 {
			cancel()
		}
	}

	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}
	if !events[0].Changed {
		t.Error("first event should be changed (first seen)")
	}
	if !events[1].Changed {
		t.Error("second event should be changed (value updated)")
	}
}

func TestSecretDataEqual(t *testing.T) {
	a := map[string]interface{}{"x": "1", "y": "2"}
	b := map[string]interface{}{"x": "1", "y": "2"}
	if !secretDataEqual(a, b) {
		t.Error("expected equal maps to be equal")
	}
	b["y"] = "3"
	if secretDataEqual(a, b) {
		t.Error("expected different maps to not be equal")
	}
}
