package vault

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// MigrateOptions controls the behaviour of a single secret migration.
type MigrateOptions struct {
	DryRun   bool
	Rollback bool // register writes with a Rollbacker when non-nil
}

// Migrator moves secrets from a source Vault client to a destination.
type Migrator struct {
	src        *Client
	dst        *Client
	logger     *audit.Logger
	rollbacker *Rollbacker
}

// NewMigrator constructs a Migrator. Pass a non-nil Rollbacker to enable
// automatic rollback registration on each successful write.
func NewMigrator(src, dst *Client, l *audit.Logger, rb *Rollbacker) *Migrator {
	return &Migrator{src: src, dst: dst, logger: l, rollbacker: rb}
}

// Migrate reads the secret at srcPath from the source Vault and writes it to
// dstPath on the destination Vault.
func (m *Migrator) Migrate(ctx context.Context, srcMount, srcPath, dstMount, dstPath string, opts MigrateOptions) error {
	data, err := m.src.ReadSecret(ctx, srcMount, srcPath)
	if err != nil {
		m.logger.Log("read_error", map[string]interface{}{
			"src_path": srcPath,
			"error":    err.Error(),
		})
		return fmt.Errorf("read %s: %w", srcPath, err)
	}

	m.logger.Log("read_ok", map[string]interface{}{"src_path": srcPath})

	if opts.DryRun {
		m.logger.Log("dry_run_skip", map[string]interface{}{"dst_path": dstPath})
		return nil
	}

	if err := m.dst.WriteSecret(ctx, dstMount, dstPath, data); err != nil {
		m.logger.Log("write_error", map[string]interface{}{
			"dst_path": dstPath,
			"error":    err.Error(),
		})
		return fmt.Errorf("write %s: %w", dstPath, err)
	}

	m.logger.Log("write_ok", map[string]interface{}{"dst_path": dstPath})

	if m.rollbacker != nil {
		m.rollbacker.Record(dstMount, dstPath)
	}
	return nil
}

// Errors accumulates non-fatal migration errors across a batch run.
type Errors []error

func (e Errors) Error() string {
	return fmt.Sprintf("%d migration error(s)", len(e))
}
