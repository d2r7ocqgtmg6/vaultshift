package vault

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newLockMockServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	store := &sync.Map{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if val, ok := store.Load(r.URL.Path); ok {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"data":{"owner":"` + val.(string) + `"}}`)) //nolint
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPut, http.MethodPost:
			store.Store(r.URL.Path, "test-agent")
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			store.Delete(r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(ts.Close)
	return ts, store
}

func newLockClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestNewLocker_MissingClient(t *testing.T) {
	_, err := NewLocker(nil, "secret/lock", "agent-1")
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewLocker_MissingPath(t *testing.T) {
	ts, _ := newLockMockServer(t)
	c := newLockClient(t, ts.URL)
	_, err := NewLocker(c, "", "agent-1")
	if err == nil {
		t.Fatal("expected error for empty lockPath")
	}
}

func TestLocker_AcquireAndRelease(t *testing.T) {
	ts, _ := newLockMockServer(t)
	c := newLockClient(t, ts.URL)

	locker, err := NewLocker(c, "secret/data/locks/migrate", "agent-1")
	if err != nil {
		t.Fatalf("NewLocker: %v", err)
	}

	if err := locker.Acquire(); err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	locked, err := locker.IsLocked()
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if !locked {
		t.Fatal("expected lock to be held after Acquire")
	}

	if err := locker.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

func TestLocker_IsLocked_WhenFree(t *testing.T) {
	ts, _ := newLockMockServer(t)
	c := newLockClient(t, ts.URL)

	locker, _ := NewLocker(c, "secret/data/locks/free", "agent-2")
	locked, err := locker.IsLocked()
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if locked {
		t.Fatal("expected lock to be free on a fresh path")
	}
}
