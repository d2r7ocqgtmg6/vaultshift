package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newClassifyMockServer(secrets map[string]map[string]any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		data, ok := secrets[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
}

func newClassifyClient(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := New(addr, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newClassifyLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestNewClassifier_MissingClient(t *testing.T) {
	l := newClassifyLogger(t)
	_, err := NewClassifier(nil, l, map[string]string{"password": "sensitive"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClassifier_NoRules(t *testing.T) {
	svr := newClassifyMockServer(nil)
	defer svr.Close()
	c := newClassifyClient(t, svr.URL)
	l := newClassifyLogger(t)
	_, err := NewClassifier(c, l, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestClassify_MatchesRule(t *testing.T) {
	svr := newClassifyMockServer(map[string]map[string]any{
		"secret/app": {"db_password": "s3cr3t"},
	})
	defer svr.Close()
	c := newClassifyClient(t, svr.URL)
	l := newClassifyLogger(t)
	cl, _ := NewClassifier(c, l, map[string]string{"password": "sensitive"})
	res, err := cl.Classify("secret/app")
	if err != nil {
		t.Fatal(err)
	}
	if res.Label != "sensitive" {
		t.Errorf("expected sensitive, got %s", res.Label)
	}
}

func TestClassify_Unclassified(t *testing.T) {
	svr := newClassifyMockServer(map[string]map[string]any{
		"secret/app": {"region": "us-east-1"},
	})
	defer svr.Close()
	c := newClassifyClient(t, svr.URL)
	l := newClassifyLogger(t)
	cl, _ := NewClassifier(c, l, map[string]string{"password": "sensitive"})
	res, err := cl.Classify("secret/app")
	if err != nil {
		t.Fatal(err)
	}
	if res.Label != "unclassified" {
		t.Errorf("expected unclassified, got %s", res.Label)
	}
}
