package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// EventType represents the type of audit event.
type EventType string

const (
	EventRead   EventType = "READ"
	EventWrite  EventType = "WRITE"
	EventDelete EventType = "DELETE"
	EventSkip   EventType = "SKIP"
	EventError  EventType = "ERROR"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Event     EventType `json:"event"`
	Path      string    `json:"path"`
	Namespace string    `json:"namespace"`
	DryRun    bool      `json:"dry_run"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Logger writes structured audit entries to a destination.
type Logger struct {
	writer io.Writer
	dryRun bool
}

// New creates a new Logger. If logPath is empty, logs are written to stdout.
func New(logPath string, dryRun bool) (*Logger, error) {
	var w io.Writer = os.Stdout
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return nil, fmt.Errorf("audit: opening log file %q: %w", logPath, err)
		}
		w = f
	}
	return &Logger{writer: w, dryRun: dryRun}, nil
}

// Log writes an audit entry for the given event.
func (l *Logger) Log(event EventType, namespace, path, message, errMsg string) error {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Event:     event,
		Path:      path,
		Namespace: namespace,
		DryRun:    l.dryRun,
		Message:   message,
		Error:     errMsg,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshalling entry: %w", err)
	}
	_, err = fmt.Fprintf(l.writer, "%s\n", data)
	return err
}
