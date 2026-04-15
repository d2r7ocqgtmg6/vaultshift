package vault_test

import (
	"os"
	"testing"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/vault"
)

func newTestLogger(t *testing.T) *audit.Logger {
	t.Helper()
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("creating logger: %v", err)
	}
	return l
}

func TestMigrate_DryRun_SkipsWrite(t *testing.T) {
	// Use a server that returns a valid secret for reads.
	readHandler := mockSecretHandler(t)
	defer readHandler.Close()

	src, _ := vault.New(vault.Config{Address: readHandler.URL, Token: "root"})
	dst, _ := vault.New(vault.Config{Address: "http://127.0.0.1:1", Token: "root"})

	m := vault.NewMigrator(src, dst, newTestLogger(t))
	result, err := m.Migrate(vault.MigrateOptions{
		SourcePath: "secret/data/app/db",
		DestPath:   "secret/data/prod/db",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
	if len(result.Migrated) != 0 {
		t.Errorf("expected 0 migrated in dry-run, got %d", len(result.Migrated))
	}
}

func TestMigrate_WriteError_RecordedInErrors(t *testing.T) {
	readHandler := mockSecretHandler(t)
	defer readHandler.Close()

	src, _ := vault.New(vault.Config{Address: readHandler.URL, Token: "root"})
	// Destination points to unreachable address to force write error.
	dst, _ := vault.New(vault.Config{Address: "http://127.0.0.1:1", Token: "root"})

	m := vault.NewMigrator(src, dst, newTestLogger(t))
	result, _ := m.Migrate(vault.MigrateOptions{
		SourcePath: "secret/data/app/db",
		DestPath:   "secret/data/prod/db",
		DryRun:     false,
	})
	if len(result.Errors) == 0 {
		t.Error("expected errors when destination is unreachable")
	}
}

// mockSecretHandler returns a test server that responds with a minimal KV secret.
func mockSecretHandler(t *testing.T) *httptest.Server {
	t.Helper()
	return newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"password":"s3cr3t"}}`))
	})
}

func init() {
	// Silence stdout during tests.
	os.Stdout, _ = os.Open(os.DevNull)
}
