package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	l, err := New("", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_FileCreation(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "audit.log")
	l, err := New(tmp, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	if _, statErr := os.Stat(tmp); statErr != nil {
		t.Fatalf("expected log file to exist: %v", statErr)
	}
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/dir/audit.log", false)
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestLog_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{writer: &buf, dryRun: true}

	if err := l.Log(EventWrite, "ns1", "secret/foo", "migrated", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry Entry
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if entry.Event != EventWrite {
		t.Errorf("expected event %q, got %q", EventWrite, entry.Event)
	}
	if entry.Namespace != "ns1" {
		t.Errorf("expected namespace %q, got %q", "ns1", entry.Namespace)
	}
	if entry.Path != "secret/foo" {
		t.Errorf("expected path %q, got %q", "secret/foo", entry.Path)
	}
	if !entry.DryRun {
		t.Error("expected dry_run to be true")
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLog_ErrorField(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{writer: &buf, dryRun: false}

	if err := l.Log(EventError, "ns2", "secret/bar", "", "permission denied"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if entry.Error != "permission denied" {
		t.Errorf("expected error field %q, got %q", "permission denied", entry.Error)
	}
}
