package vault

import (
	"fmt"
	"io"
	"sort"

	"github.com/your-org/vaultshift/internal/audit"
)

// AuditSummary aggregates counts of audit log events by operation and status.
type AuditSummary struct {
	Total    int
	ByOp     map[string]int
	ByStatus map[string]int
	Errors   []string
}

// AuditSummarizer reads audit log entries and produces a summary.
type AuditSummarizer struct {
	logger *audit.Logger
}

// NewAuditSummarizer creates an AuditSummarizer. Returns an error if logger is nil.
func NewAuditSummarizer(logger *audit.Logger) (*AuditSummarizer, error) {
	if logger == nil {
		return nil, fmt.Errorf("audit summarizer: logger is required")
	}
	return &AuditSummarizer{logger: logger}, nil
}

// Summarize reads all entries from the given reader (JSONL) and returns an AuditSummary.
func (s *AuditSummarizer) Summarize(r io.Reader) (*AuditSummary, error) {
	entries, err := s.logger.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("audit summarizer: read entries: %w", err)
	}

	summary := &AuditSummary{
		ByOp:     make(map[string]int),
		ByStatus: make(map[string]int),
	}

	for _, e := range entries {
		summary.Total++
		if e.Op != "" {
			summary.ByOp[e.Op]++
		}
		if e.Status != "" {
			summary.ByStatus[e.Status]++
		}
		if e.Error != "" {
			summary.Errors = append(summary.Errors, e.Error)
		}
	}
	return summary, nil
}

// Print writes a human-readable summary to w.
func (s *AuditSummarizer) Print(w io.Writer, summary *AuditSummary) {
	fmt.Fprintf(w, "Total events: %d\n", summary.Total)

	ops := sortedKeys(summary.ByOp)
	if len(ops) > 0 {
		fmt.Fprintln(w, "By operation:")
		for _, op := range ops {
			fmt.Fprintf(w, "  %-20s %d\n", op, summary.ByOp[op])
		}
	}

	statuses := sortedKeys(summary.ByStatus)
	if len(statuses) > 0 {
		fmt.Fprintln(w, "By status:")
		for _, st := range statuses {
			fmt.Fprintf(w, "  %-20s %d\n", st, summary.ByStatus[st])
		}
	}

	if len(summary.Errors) > 0 {
		fmt.Fprintf(w, "Errors (%d):\n", len(summary.Errors))
		for _, e := range summary.Errors {
			fmt.Fprintf(w, "  - %s\n", e)
		}
	}
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
