package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [snapshot-file]",
	Short: "Restore secrets from a snapshot file to a destination prefix",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestore,
}

func init() {
	restoreCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	restoreCmd.Flags().String("dest-prefix", "", "Destination prefix to restore secrets into")
	restoreCmd.Flags().Bool("dry-run", false, "Preview restore without writing")
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	destPrefix, _ := cmd.Flags().GetString("dest-prefix")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if destPrefix == "" {
		return fmt.Errorf("--dest-prefix is required")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	snap, err := vault.LoadSnapshot(args[0])
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	restorer, err := vault.NewRestorer(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("restorer: %w", err)
	}

	result, err := restorer.Restore(cmd.Context(), snap, destPrefix)
	if err != nil {
		return fmt.Errorf("restore: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Restored: %d  Skipped: %d  Errors: %d\n",
		len(result.Restored), len(result.Skipped), len(result.Errors))
	for _, e := range result.Errors {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", e)
	}
	return nil
}
