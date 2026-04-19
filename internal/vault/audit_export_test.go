package vault

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultshift/internal/audit"
)

func newAuditExportLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewAuditExporter_NilLogger(t *testing.T) {
	_, err := NewAuditExporter(nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewAuditExporter_Valid(t *testing.T) {
	l := newAuditExportLogger(t)
	ex, err := NewAuditExporter(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ex == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestAuditExport_DryRun_NoFile(t *testing.T) {
	l := newAuditExportLogger(t)
	ex, _ := NewAuditExporter(l)
	entries := []AuditExportEntry{
		{Timestamp: time.Now(), Operation: "migrate", Path: "secret/foo", Status: "ok"},
	}
	dest := filepath.Join(t.TempDir(), "audit.jsonl")
	n, err := ex.Export(entries, dest, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("expected no file created in dry-run")
	}
}

func TestAuditExport_WritesJSONL(t *testing.T) {
	l := newAuditExportLogger(t)
	ex, _ := NewAuditExporter(l)
	entries := []AuditExportEntry{
		{Timestamp: time.Now(), Operation: "migrate", Path: "secret/a", Status: "ok"},
		{Timestamp: time.Now(), Operation: "rollback", Path: "secret/b", Status: "error"},
	}
	dest := filepath.Join(t.TempDir(), "audit.jsonl")
	n, err := ex.Export(entries, dest, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2, got %d", n)
	}
	f, _ := os.Open(dest)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var count int
	for scanner.Scan() {
		var e AuditExportEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Errorf("invalid JSON line: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 lines, got %d", count)
	}
}
