package vault

import (
	"fmt"
	"strings"

	"github.com/vaultshift/internal/audit"
)

// MigrateOptions controls migration behaviour.
type MigrateOptions struct {
	SourcePath string
	DestPath   string
	DryRun     bool
	Recursive  bool
}

// MigrateResult summarises a completed migration.
type MigrateResult struct {
	Migrated []string
	Skipped  []string
	Errors   []string
}

// Migrator moves secrets from a source Vault client to a destination.
type Migrator struct {
	src    *Client
	dst    *Client
	logger *audit.Logger
}

// NewMigrator creates a Migrator.
func NewMigrator(src, dst *Client, logger *audit.Logger) *Migrator {
	return &Migrator{src: src, dst: dst, logger: logger}
}

// Migrate performs the migration according to the provided options.
func (m *Migrator) Migrate(opts MigrateOptions) (*MigrateResult, error) {
	result := &MigrateResult{}

	paths, err := m.resolvePaths(opts)
	if err != nil {
		return nil, fmt.Errorf("resolving paths: %w", err)
	}

	for _, path := range paths {
		destPath := strings.Replace(path, opts.SourcePath, opts.DestPath, 1)

		data, err := m.src.ReadSecret(path)
		if err != nil {
			m.logger.Log("read_error", map[string]interface{}{"path": path, "error": err.Error()})
			result.Errors = append(result.Errors, path)
			continue
		}

		if opts.DryRun {
			m.logger.Log("dry_run", map[string]interface{}{"src": path, "dst": destPath})
			result.Skipped = append(result.Skipped, path)
			continue
		}

		if err := m.dst.WriteSecret(destPath, data); err != nil {
			m.logger.Log("write_error", map[string]interface{}{"path": destPath, "error": err.Error()})
			result.Errors = append(result.Errors, path)
			continue
		}

		m.logger.Log("migrated", map[string]interface{}{"src": path, "dst": destPath})
		result.Migrated = append(result.Migrated, path)
	}

	return result, nil
}

func (m *Migrator) resolvePaths(opts MigrateOptions) ([]string, error) {
	if !opts.Recursive {
		return []string{opts.SourcePath}, nil
	}
	keys, err := m.src.ListSecrets(opts.SourcePath)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(keys))
	for _, k := range keys {
		if !strings.HasSuffix(k, "/") {
			paths = append(paths, opts.SourcePath+"/"+k)
		}
	}
	return paths, nil
}
