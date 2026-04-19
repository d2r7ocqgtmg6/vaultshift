package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newPolicyMockServer(t *testing.T, caps map[string][]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/capabilities-self" {
			data := map[string]interface{}{}
			for k, v := range caps {
				iface := make([]interface{}, len(v))
				for i, s := range v {
					iface[i] = s
				}
				data[k] = iface
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newPolicyLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestNewPolicyChecker_MissingClient(t *testing.T) {
	_, err := NewPolicyChecker(nil, newPolicyLogger(t))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewPolicyChecker_MissingLogger(t *testing.T) {
	client := &Client{}
	_, err := NewPolicyChecker(client, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPolicyCheck_AllowedPath(t *testing.T) {
	srv := newPolicyMockServer(t, map[string][]string{
		"secret/data/foo": {"read", "list"},
	})
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	checker, err := NewPolicyChecker(client, newPolicyLogger(t))
	if err != nil {
		t.Fatal(err)
	}

	results, err := checker.Check([]string{"secret/data/foo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Allowed {
		t.Error("expected path to be allowed")
	}
}

func TestPolicyCheck_NoPaths(t *testing.T) {
	client := &Client{}
	checker, _ := NewPolicyChecker(client, newPolicyLogger(t))
	_, err := checker.Check([]string{})
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}
