package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// AuditTrailer writes a summary of migration actions to the audit log.
type AuditTrailer struct {
	logger *audit.Logger
}

// NewAuditTrailer creates a new AuditTrailer.
func NewAuditTrailer(logger *audit.Logger) (*AuditTrailer, error) {
	if logger == nil {
		return nil, fmt.Errorf("audit logger is required")
	}
	return &AuditTrailer{logger: logger}, nil
}

// RecordMigration logs a single secret migration event.
func (a *AuditTrailer) RecordMigration(path, srcNamespace, dstNamespace string, dryRun bool) {
	a.logger.Log(map[string]interface{}{
		"event":         "migrate",
		"path":          path,
		"src_namespace": srcNamespace,
		"dst_namespace": dstNamespace,
		"dry_run":       dryRun,
	})
}

// RecordError logs a failed migration event.
func (a *AuditTrailer) RecordError(path string, err error) {
	a.logger.Log(map[string]interface{}{
		"event": "migrate_error",
		"path":  path,
		"error": err.Error(),
	})
}

// RecordSummary logs a final summary of the migration run.
func (a *AuditTrailer) RecordSummary(total, succeeded, failed int, dryRun bool) {
	a.logger.Log(map[string]interface{}{
		"event":     "summary",
		"total":     total,
		"succeeded": succeeded,
		"failed":    failed,
		"dry_run":   dryRun,
	})
}
