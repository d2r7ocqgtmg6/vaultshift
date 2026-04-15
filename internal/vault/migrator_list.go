package vault

import (
	"context"
	"fmt"
)

// MigrateAll discovers all secret paths under srcPrefix and migrates each one.
// It returns the total count of successfully migrated secrets and any errors encountered.
func (m *Migrator) MigrateAll(ctx context.Context, srcPrefix string) (int, []error) {
	paths, err := m.src.ListSecrets(ctx, srcPrefix)
	if err != nil {
		return 0, []error{fmt.Errorf("listing source secrets: %w", err)}
	}

	if len(paths) == 0 {
		m.logger.Log("info", "no secrets found under prefix", map[string]string{
			"prefix": srcPrefix,
		})
		return 0, nil
	}

	var errs []error
	successCount := 0

	for _, path := range paths {
		if migrateErr := m.Migrate(ctx, path, path); migrateErr != nil {
			errs = append(errs, fmt.Errorf("path %q: %w", path, migrateErr))
			continue
		}
		successCount++
	}

	m.logger.Log("info", "migration complete", map[string]string{
		"prefix":         srcPrefix,
		"total":          fmt.Sprintf("%d", len(paths)),
		"success":        fmt.Sprintf("%d", successCount),
		"error_count":    fmt.Sprintf("%d", len(errs)),
	})

	return successCount, errs
}
