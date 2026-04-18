package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newValidateMockServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if data == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func newValidateClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func newValidateLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewValidator_MissingClient(t *testing.T) {
	l := newValidateLogger(t)
	_, err := NewValidator(nil, l, []ValidationRule{RequiredKeys("x")}, false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewValidator_NoRules(t *testing.T) {
	srv := newValidateMockServer(map[string]interface{}{"k": "v"})
	defer srv.Close()
	c := newValidateClient(t, srv)
	l := newValidateLogger(t)
	_, err := NewValidator(c, l, nil, false)
	if err == nil {
		t.Fatal("expected error for no rules")
	}
}

func TestValidate_RequiredKeys_Pass(t *testing.T) {
	srv := newValidateMockServer(map[string]interface{}{"username": "admin", "password": "s3cr3t"})
	defer srv.Close()
	c := newValidateClient(t, srv)
	l := newValidateLogger(t)
	v, _ := NewValidator(c, l, []ValidationRule{RequiredKeys("username", "password")}, false)
	results, err := v.Validate("secret/app")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("expected pass, got %+v", results)
	}
}

func TestValidate_RequiredKeys_Fail(t *testing.T) {
	srv := newValidateMockServer(map[string]interface{}{"username": "admin"})
	defer srv.Close()
	c := newValidateClient(t, srv)
	l := newValidateLogger(t)
	v, _ := NewValidator(c, l, []ValidationRule{RequiredKeys("username", "password")}, false)
	results, err := v.Validate("secret/app")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if results[0].Passed {
		t.Fatal("expected fail")
	}
}

func TestValidate_NoEmptyValues_Fail(t *testing.T) {
	srv := newValidateMockServer(map[string]interface{}{"key": ""})
	defer srv.Close()
	c := newValidateClient(t, srv)
	l := newValidateLogger(t)
	v, _ := NewValidator(c, l, []ValidationRule{NoEmptyValues()}, false)
	results, err := v.Validate("secret/app")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if results[0].Passed {
		t.Fatal("expected fail for empty value")
	}
}
