package vault

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/vaultshift/internal/audit"
)

func newAuditSummaryLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("audit.New: %v", err)
	}
	return l
}

func TestNewAuditSummarizer_NilLogger(t *testing.T) {
	_, err := NewAuditSummarizer(nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewAuditSummarizer_Valid(t *testing.T) {
	l := newAuditSummaryLogger(t)
	s, err := NewAuditSummarizer(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil summarizer")
	}
}

func TestAuditSummarize_CountsByOpAndStatus(t *testing.T) {
	l := newAuditSummaryLogger(t)
	s, _ := NewAuditSummarizer(l)

	input := strings.NewReader(
		`{"op":"write","status":"ok","error":""}` + "\n" +
			`{"op":"write","status":"ok","error":""}` + "\n" +
			`{"op":"delete","status":"error","error":"permission denied"}` + "\n",
	)

	summary, err := s.Summarize(input)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}

	if summary.Total != 3 {
		t.Errorf("expected Total=3, got %d", summary.Total)
	}
	if summary.ByOp["write"] != 2 {
		t.Errorf("expected ByOp[write]=2, got %d", summary.ByOp["write"])
	}
	if summary.ByOp["delete"] != 1 {
		t.Errorf("expected ByOp[delete]=1, got %d", summary.ByOp["delete"])
	}
	if summary.ByStatus["ok"] != 2 {
		t.Errorf("expected ByStatus[ok]=2, got %d", summary.ByStatus["ok"])
	}
	if len(summary.Errors) != 1 || summary.Errors[0] != "permission denied" {
		t.Errorf("unexpected errors: %v", summary.Errors)
	}
}

func TestAuditSummarize_EmptyInput(t *testing.T) {
	l := newAuditSummaryLogger(t)
	s, _ := NewAuditSummarizer(l)

	summary, err := s.Summarize(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 0 {
		t.Errorf("expected Total=0, got %d", summary.Total)
	}
}

func TestAuditSummarize_Print(t *testing.T) {
	l := newAuditSummaryLogger(t)
	s, _ := NewAuditSummarizer(l)

	summary := &AuditSummary{
		Total:    2,
		ByOp:     map[string]int{"write": 2},
		ByStatus: map[string]int{"ok": 2},
		Errors:   nil,
	}

	var buf bytes.Buffer
	s.Print(&buf, summary)
	out := buf.String()

	if !strings.Contains(out, "Total events: 2") {
		t.Errorf("expected total in output, got: %s", out)
	}
	if !strings.Contains(out, "write") {
		t.Errorf("expected op 'write' in output, got: %s", out)
	}
}
