package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newAlignLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func newAlignMockServer(secrets map[string]map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			if data, ok := secrets[path]; ok {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				secrets[path] = d.(map[string]interface{})
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			delete(secrets, path)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestNewAligner_MissingSource(t *testing.T) {
	l := newAlignLogger(t)
	_, err := NewAligner(nil, &Client{}, l, false, false)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNewAligner_MissingDest(t *testing.T) {
	l := newAlignLogger(t)
	_, err := NewAligner(&Client{}, nil, l, false, false)
	if err == nil {
		t.Fatal("expected error for nil dest")
	}
}

func TestAlign_DryRun_NoWrite(t *testing.T) {
	srcSecrets := map[string]map[string]interface{}{
		"/v1/secret/data/app/key": {"password": "s3cr3t"},
	}
	dstSecrets := map[string]map[string]interface{}{}

	srcSrv := newAlignMockServer(srcSecrets)
	dstSrv := newAlignMockServer(dstSecrets)
	defer srcSrv.Close()
	defer dstSrv.Close()

	src, _ := New(srcSrv.URL, "tok", "")
	dst, _ := New(dstSrv.URL, "tok", "")
	l := newAlignLogger(t)

	a, err := NewAligner(src, dst, l, true, false)
	if err != nil {
		t.Fatalf("NewAligner: %v", err)
	}

	result, err := a.Align("secret/app")
	if err != nil {
		t.Fatalf("Align: %v", err)
	}
	if len(result.Written) == 0 {
		t.Error("expected dry-run written entry")
	}
	if len(dstSecrets) != 0 {
		t.Error("dry-run should not write to destination")
	}
}

func TestAlign_Prune_RemovesOrphaned(t *testing.T) {
	srcSecrets := map[string]map[string]interface{}{}
	dstSecrets := map[string]map[string]interface{}{
		"/v1/secret/data/app/old": {"key": "val"},
	}

	srcSrv := newAlignMockServer(srcSecrets)
	dstSrv := newAlignMockServer(dstSecrets)
	defer srcSrv.Close()
	defer dstSrv.Close()

	src, _ := New(srcSrv.URL, "tok", "")
	dst, _ := New(dstSrv.URL, "tok", "")
	l := newAlignLogger(t)

	a, err := NewAligner(src, dst, l, false, true)
	if err != nil {
		t.Fatalf("NewAligner: %v", err)
	}

	result, err := a.Align("secret/app")
	if err != nil {
		t.Fatalf("Align: %v", err)
	}
	if len(result.Pruned) == 0 {
		t.Error("expected pruned entry")
	}
}
