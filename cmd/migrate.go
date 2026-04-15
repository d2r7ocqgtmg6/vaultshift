package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate secrets from source to destination Vault",
	Long:  `Recursively migrates secrets from a source Vault namespace/prefix to a destination Vault namespace/prefix.`,
	RunE:  runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if dryRun {
		cfg.DryRun = true
	}

	logPath := auditLog
	if logPath == "" {
		logPath = cfg.AuditLog
	}

	logger, err := audit.New(logPath)
	if err != nil {
		return fmt.Errorf("initialising audit logger: %w", err)
	}

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token, cfg.Source.Namespace)
	if err != nil {
		return fmt.Errorf("creating source vault client: %w", err)
	}

	dst, err := vault.New(cfg.Destination.Address, cfg.Destination.Token, cfg.Destination.Namespace)
	if err != nil {
		return fmt.Errorf("creating destination vault client: %w", err)
	}

	migrator := vault.NewMigrator(src, dst, logger, cfg.DryRun)

	result, err := migrator.MigrateAll(cmd.Context(), cfg.Source.Prefix, cfg.Destination.Prefix)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Migration complete: %d migrated, %d skipped, %d errors\n",
		result.Migrated, result.Skipped, len(result.Errors))

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  error: %s\n", e)
		}
		return fmt.Errorf("%d secret(s) failed to migrate", len(result.Errors))
	}

	return nil
}
