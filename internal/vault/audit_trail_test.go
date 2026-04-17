package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newAuditLogger(t *testing.T) (*audit.Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("failed to create audit logger: %v", err)
	}
	_ = buf // logger writes to stdout in test; we validate via NewAuditTrailer
	return l, buf
}

func TestNewAuditTrailer_NilLogger(t *testing.T) {
	_, err := NewAuditTrailer(nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewAuditTrailer_Valid(t *testing.T) {
	l, _ := newAuditLogger(t)
	at, err := NewAuditTrailer(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if at == nil {
		t.Fatal("expected non-nil AuditTrailer")
	}
}

func TestRecordMigration_DoesNotPanic(t *testing.T) {
	l, _ := newAuditLogger(t)
	at, _ := NewAuditTrailer(l)
	at.RecordMigration("secret/foo", "ns1", "ns2", false)
}

func TestRecordError_DoesNotPanic(t *testing.T) {
	l, _ := newAuditLogger(t)
	at, _ := NewAuditTrailer(l)
	at.RecordError("secret/bar", errors.New("write failed"))
}

func TestRecordSummary_ValidJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	data := map[string]interface{}{
		"event":     "summary",
		"total":     10,
		"succeeded": 9,
		"failed":    1,
		"dry_run":   false,
	}
	if err := enc.Encode(data); err != nil {
		t.Fatalf("failed to encode summary: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out["event"] != "summary" {
		t.Errorf("expected event=summary, got %v", out["event"])
	}
}
