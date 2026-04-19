package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/vaultshift/internal/audit"
)

// AuditExportEntry represents a single exported audit record.
type AuditExportEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Operation string            `json:"operation"`
	Path      string            `json:"path"`
	Status    string            `json:"status"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// AuditExporter reads audit log entries and writes them to a destination file.
type AuditExporter struct {
	logger *audit.Logger
}

// NewAuditExporter creates a new AuditExporter.
func NewAuditExporter(logger *audit.Logger) (*AuditExporter, error) {
	if logger == nil {
		return nil, fmt.Errorf("audit exporter: logger is required")
	}
	return &AuditExporter{logger: logger}, nil
}

// Export writes the given entries as JSON lines to destPath.
func (e *AuditExporter) Export(entries []AuditExportEntry, destPath string, dryRun bool) (int, error) {
	if dryRun {
		e.logger.Log("audit_export", "dry-run: would export entries", map[string]interface{}{
			"count": len(entries),
			"dest":  destPath,
		})
		return len(entries), nil
	}

	f, err := os.Create(destPath)
	if err != nil {
		return 0, fmt.Errorf("audit exporter: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, entry := range entries {
		if err := enc.Encode(entry); err != nil {
			return 0, fmt.Errorf("audit exporter: encode entry: %w", err)
		}
	}

	e.logger.Log("audit_export", "exported audit entries", map[string]interface{}{
		"count": len(entries),
		"dest":  destPath,
	})
	return len(entries), nil
}
