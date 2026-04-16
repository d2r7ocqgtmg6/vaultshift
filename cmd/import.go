package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import secrets from a JSON file into Vault",
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func init() {
	importCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	importCmd.Flags().Bool("dry-run", false, "Preview import without writing")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	filePath := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	client, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("init vault client: %w", err)
	}

	imp, err := vault.NewImporter(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("init importer: %w", err)
	}

	n, err := imp.Import(cmd.Context(), filePath)
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	fmt.Fprintf(os.Stdout, "imported %d secret(s) from %s (dry-run=%v)\n", n, filePath, dryRun)
	return nil
}
