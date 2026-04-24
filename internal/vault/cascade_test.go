package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newCascadeLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newCascadeMockServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			data, ok := secrets[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if inner, ok := body["data"]; ok {
				if m, ok := inner.(map[string]interface{}); ok {
					secrets[path] = m
				}
			} else {
				secrets[path] = body
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestNewCascader_MissingClient(t *testing.T) {
	_, err := NewCascader(nil, newCascadeLogger(t), false, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewCascader_MissingLogger(t *testing.T) {
	client, _ := New("http://127.0.0.1", "tok")
	_, err := NewCascader(client, nil, false, false)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestCascade_DryRun_NoWrite(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/src": {"key": "value"},
	}
	srv := newCascadeMockServer(t, secrets)
	defer srv.Close()

	client, _ := New(srv.URL, "tok")
	cascader, _ := NewCascader(client, newCascadeLogger(t), true, false)

	results, err := cascader.Cascade("secret/src", []string{"secret/dst"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Written {
		t.Error("dry-run should not write")
	}
	if _, written := secrets["secret/dst"]; written {
		t.Error("dry-run must not mutate secrets map")
	}
}

func TestCascade_WritesPayload(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/src": {"db_pass": "s3cr3t", "api_key": "abc123"},
	}
	srv := newCascadeMockServer(t, secrets)
	defer srv.Close()

	client, _ := New(srv.URL, "tok")
	cascader, _ := NewCascader(client, newCascadeLogger(t), false, true)

	results, err := cascader.Cascade("secret/src", []string{"secret/dst"}, []string{"db_pass"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results[0].Written {
		t.Error("expected write to succeed")
	}
	if secrets["secret/dst"]["db_pass"] != "s3cr3t" {
		t.Errorf("expected db_pass to be cascaded, got %v", secrets["secret/dst"])
	}
	if _, present := secrets["secret/dst"]["api_key"]; present {
		t.Error("api_key should not have been cascaded")
	}
}

func TestCascade_SkipsExistingWithoutOverwrite(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/src": {"key": "new_value"},
		"secret/dst": {"key": "old_value"},
	}
	srv := newCascadeMockServer(t, secrets)
	defer srv.Close()

	client, _ := New(srv.URL, "tok")
	cascader, _ := NewCascader(client, newCascadeLogger(t), false, false)

	results, err := cascader.Cascade("secret/src", []string{"secret/dst"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Skipped != 1 {
		t.Errorf("expected 1 skipped key, got %d", results[0].Skipped)
	}
	if secrets["secret/dst"]["key"] != "old_value" {
		t.Error("existing key should not be overwritten")
	}
}
