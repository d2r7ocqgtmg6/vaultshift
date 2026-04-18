package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultshift/internal/audit"
)

func newQuotaMockServer(keys []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Query().Get("list") == "true" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": keys},
			})
			return
		}
		http.NotFound(w, r)
	}))
}

func newQuotaClient(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := New(addr, "token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewQuoter_MissingClient(t *testing.T) {
	_, err := NewQuoter(nil, nil, 10)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewQuoter_InvalidLimit(t *testing.T) {
	srv := newQuotaMockServer(nil)
	defer srv.Close()
	c := newQuotaClient(t, srv.URL)
	_, err := NewQuoter(c, nil, 0)
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestQuota_BelowLimit(t *testing.T) {
	srv := newQuotaMockServer([]string{"a", "b"})
	defer srv.Close()
	c := newQuotaClient(t, srv.URL)
	logger, _ := audit.New("")
	q, _ := NewQuoter(c, logger, 5)
	results, err := q.Check([]string{"secret/"})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Exceeds {
		t.Error("expected quota not exceeded")
	}
	if results[0].Count != 2 {
		t.Errorf("expected count 2, got %d", results[0].Count)
	}
}

func TestQuota_ExceedsLimit(t *testing.T) {
	srv := newQuotaMockServer([]string{"a", "b", "c"})
	defer srv.Close()
	c := newQuotaClient(t, srv.URL)
	q, _ := NewQuoter(c, nil, 2)
	results, err := q.Check([]string{"secret/"})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !results[0].Exceeds {
		t.Error("expected quota exceeded")
	}
}
