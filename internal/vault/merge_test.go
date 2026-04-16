package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
)

func newMergeLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newMergeMockServer(t *testing.T, secrets map[string]map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		if r.Method == http.MethodGet {
			if data, ok := secrets[path]; ok {
				json.NewEncoder(w).Encode(map[string]any{"data": data})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			secrets[path] = body
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestNewMerger_MissingClient(t *testing.T) {
	l := newMergeLogger(t)
	_, err := NewMerger(nil, nil, l, false, false)
	if err == nil {
		t.Fatal("expected error for nil src client")
	}
}

func TestMerge_DryRun_NoWrite(t *testing.T) {
	srcSecrets := map[string]map[string]any{"secret/src/key": {"value": "abc"}}
	dstSecrets := map[string]map[string]any{}

	srcSrv := newMergeMockServer(t, srcSecrets)
	dstSrv := newMergeMockServer(t, dstSecrets)
	defer srcSrv.Close()
	defer dstSrv.Close()

	src, _ := New(srcSrv.URL, "tok", "")
	dst, _ := New(dstSrv.URL, "tok", "")
	l := newMergeLogger(t)

	m, err := NewMerger(src, dst, l, true, false)
	if err != nil {
		t.Fatalf("NewMerger: %v", err)
	}

	results, err := m.Merge("secret/src/", "secret/dst/")
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(dstSecrets) != 0 {
		t.Error("expected no writes in dry-run mode")
	}
	_ = results
}

func TestMerge_SkipsExistingWithoutOverwrite(t *testing.T) {
	srcSecrets := map[string]map[string]any{"secret/src/key": {"value": "new"}}
	dstSecrets := map[string]map[string]any{"secret/dst/key": {"value": "old"}}

	srcSrv := newMergeMockServer(t, srcSecrets)
	dstSrv := newMergeMockServer(t, dstSecrets)
	defer srcSrv.Close()
	defer dstSrv.Close()

	src, _ := New(srcSrv.URL, "tok", "")
	dst, _ := New(dstSrv.URL, "tok", "")
	l := newMergeLogger(t)

	m, _ := NewMerger(src, dst, l, false, false)
	results, err := m.Merge("secret/src/", "secret/dst/")
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(results) == 0 || results[0].Action != "skipped" {
		t.Errorf("expected skipped, got %+v", results)
	}
}
