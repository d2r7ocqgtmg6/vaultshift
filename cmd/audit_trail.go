package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var auditTrailCmd = &cobra.Command{
	Use:   "audit-trail",
	Short: "Record a manual audit trail entry for a secret path",
	Args:  cobra.ExactArgs(1),
	RunE:  runAuditTrail,
}

func init() {
	auditTrailCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	auditTrailCmd.Flags().String("log", "", "Audit log output file (default: stdout)")
	auditTrailCmd.Flags().Bool("dry-run", false, "Mark entry as dry-run")
	rootCmd.AddCommand(auditTrailCmd)
}

func runAuditTrail(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	logPath, _ := cmd.Flags().GetString("log")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	l, err := audit.New(logPath)
	if err != nil {
		return fmt.Errorf("create audit logger: %w", err)
	}

	at, err := vault.NewAuditTrailer(l)
	if err != nil {
		return fmt.Errorf("create audit trailer: %w", err)
	}

	path := args[0]
	at.RecordMigration(path, cfg.Source.Namespace, cfg.Destination.Namespace, dryRun)
	fmt.Fprintf(cmd.OutOrStdout(), "audit entry recorded for path: %s\n", path)
	return nil
}
