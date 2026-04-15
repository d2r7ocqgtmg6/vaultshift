package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newExpireMockServer(t *testing.T, secrets map[string]map[string]interface{}, deleted *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method == http.MethodDelete {
			*deleted = append(*deleted, path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		data, ok := secrets[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
	}))
}

func newExpireLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newExpireClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	return c
}

func TestExpire_NotExpired_NoDelete(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	deleted := &[]string{}
	srv := newExpireMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/svc/key": {"expires_at": future, "value": "abc"},
	}, deleted)
	defer srv.Close()

	client := newExpireClient(t, srv)
	expirer := NewExpirer(client, newExpireLogger(t), "expires_at", false)
	results := expirer.CheckAndPurge([]string{"secret/data/svc/key"})

	if len(results) != 1 || results[0].Expired {
		t.Fatalf("expected not-expired, got %+v", results)
	}
	if len(*deleted) != 0 {
		t.Fatalf("expected no deletions, got %v", *deleted)
	}
}

func TestExpire_Expired_Deletes(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	deleted := &[]string{}
	srv := newExpireMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/svc/old": {"expires_at": past, "value": "xyz"},
	}, deleted)
	defer srv.Close()

	client := newExpireClient(t, srv)
	expirer := NewExpirer(client, newExpireLogger(t), "expires_at", false)
	results := expirer.CheckAndPurge([]string{"secret/data/svc/old"})

	if len(results) != 1 || !results[0].Expired || results[0].Error != nil {
		t.Fatalf("expected expired+deleted, got %+v", results)
	}
	if len(*deleted) != 1 {
		t.Fatalf("expected 1 deletion, got %v", *deleted)
	}
}

func TestExpire_DryRun_NoDelete(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	deleted := &[]string{}
	srv := newExpireMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/svc/stale": {"expires_at": past},
	}, deleted)
	defer srv.Close()

	client := newExpireClient(t, srv)
	expirer := NewExpirer(client, newExpireLogger(t), "expires_at", true)
	results := expirer.CheckAndPurge([]string{"secret/data/svc/stale"})

	if !results[0].Expired {
		t.Fatal("expected expired=true in dry-run")
	}
	if len(*deleted) != 0 {
		t.Fatalf("dry-run should not delete, got %v", *deleted)
	}
}

func TestExpire_NoMetaKey_NotExpired(t *testing.T) {
	deleted := &[]string{}
	srv := newExpireMockServer(t, map[string]map[string]interface{}{
		"/v1/secret/data/svc/nokey": {"value": "hello"},
	}, deleted)
	defer srv.Close()

	client := newExpireClient(t, srv)
	expirer := NewExpirer(client, newExpireLogger(t), "expires_at", false)
	results := expirer.CheckAndPurge([]string{"secret/data/svc/nokey"})

	if results[0].Expired || results[0].Error != nil {
		t.Fatalf("expected not-expired, no error; got %+v", results[0])
	}
}
