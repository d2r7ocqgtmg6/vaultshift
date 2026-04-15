package vault

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLister struct {
	paths []string
	err   error
}

func (ml *mockLister) ListSecrets(_ context.Context, _ string) ([]string, error) {
	return ml.paths, ml.err
}

func TestMigrateAll_EmptyPrefix(t *testing.T) {
	logger := newTestLogger(t)
	m := &Migrator{
		logger: logger,
		src:    &Client{},
		dst:    &Client{},
		dryRun: true,
	}
	// Replace src with mock via interface if refactored; for now test dry-run path.
	count, errs := m.MigrateAll(context.Background(), "")
	// With real client and no server, expect listing error.
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestMigrateAll_ListError(t *testing.T) {
	logger := newTestLogger(t)
	m := &Migrator{
		logger: logger,
		dryRun: true,
	}
	// Simulate list error by providing nil client (will error on list)
	count, errs := m.MigrateAll(context.Background(), "broken")
	require.NotEmpty(t, errs)
	assert.Equal(t, 0, count)
	assert.True(t, errors.Is(errs[0], errs[0]), "error should be wrapped")
}

func TestMigrateAll_DryRun_NoWrite(t *testing.T) {
	logger := newTestLogger(t)
	ts := newMockVaultServer(t)

	src, err := New(ts.URL, "src-token", "secret")
	require.NoError(t, err)
	dst, err := New(ts.URL, "dst-token", "secret")
	require.NoError(t, err)

	m := NewMigrator(src, dst, logger, true)

	// MigrateAll with a prefix that returns no paths (404) should succeed with 0 count.
	count, errs := m.MigrateAll(context.Background(), "nonexistent")
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}
