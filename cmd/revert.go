package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var revertCmd = &cobra.Command{
	Use:   "revert <snapshot-file>",
	Short: "Restore secrets from a snapshot file",
	Args:  cobra.ExactArgs(1),
	RunE:  runRevert,
}

func init() {
	revertCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	revertCmd.Flags().Bool("dry-run", false, "Preview revert without writing")
	revertCmd.Flags().String("audit-log", "", "Path to audit log file")
	rootCmd.AddCommand(revertCmd)
}

func runRevert(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")
	snapshotFile := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	snap, err := vault.LoadSnapshot(snapshotFile)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	rv, err := vault.NewReverter(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("new reverter: %w", err)
	}

	results := rv.Revert(snap)
	errCount := 0
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(os.Stderr, "ERROR  %s: %v\n", r.Path, r.Err)
			errCount++
		case r.Skipped:
			fmt.Printf("DRY-RUN %s\n", r.Path)
		default:
			fmt.Printf("REVERTED %s\n", r.Path)
		}
	}
	if errCount > 0 {
		return fmt.Errorf("%d path(s) failed to revert", errCount)
	}
	return nil
}
