package vault

import (
	"fmt"
	"time"

	"github.com/vaultshift/internal/audit"
)

// TraceEntry represents a single traced operation.
type TraceEntry struct {
	Path      string        `json:"path"`
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration_ms"`
	Error     string        `json:"error,omitempty"`
}

// Tracer records operation traces for secrets paths.
type Tracer struct {
	client  *Client
	logger  *audit.Logger
	entries []TraceEntry
}

// NewTracer creates a Tracer. Returns an error if client or logger is nil.
func NewTracer(client *Client, logger *audit.Logger) (*Tracer, error) {
	if client == nil {
		return nil, fmt.Errorf("tracer: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("tracer: logger is required")
	}
	return &Tracer{client: client, logger: logger}, nil
}

// Trace reads the secret at path, measures latency, and records a TraceEntry.
func (t *Tracer) Trace(path string) TraceEntry {
	start := time.Now()
	_, err := t.client.ReadSecret(path)
	elapsed := time.Since(start)

	entry := TraceEntry{
		Path:      path,
		Operation: "read",
		Duration:  elapsed / time.Millisecond,
	}
	if err != nil {
		entry.Error = err.Error()
	}

	t.entries = append(t.entries, entry)
	t.logger.Log(map[string]interface{}{
		"op":          "trace",
		"path":        path,
		"duration_ms": entry.Duration,
		"error":       entry.Error,
	})
	return entry
}

// Entries returns all recorded trace entries.
func (t *Tracer) Entries() []TraceEntry {
	return t.entries
}
