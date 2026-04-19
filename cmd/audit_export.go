package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var auditExportCmd = &cobra.Command{
	Use:   "audit-export <dest-file>",
	Short: "Export audit log entries to a JSONL file",
	Args:  cobra.ExactArgs(1),
	RunE:  runAuditExport,
}

func init() {
	auditExportCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	auditExportCmd.Flags().Bool("dry-run", false, "preview export without writing")
	auditExportCmd.Flags().String("operation", "", "filter by operation type")
	rootCmd.AddCommand(auditExportCmd)
}

func runAuditExport(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	operation, _ := cmd.Flags().GetString("operation")
	destPath := args[0]

	_, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New("")
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	exporter, err := vault.NewAuditExporter(logger)
	if err != nil {
		return fmt.Errorf("audit exporter: %w", err)
	}

	// Placeholder entries — in production these would be read from the audit log source.
	entries := []vault.AuditExportEntry{
		{Timestamp: time.Now(), Operation: "migrate", Path: "secret/example", Status: "ok"},
	}
	if operation != "" {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Operation == operation {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	n, err := exporter.Export(entries, destPath, dryRun)
	if err != nil {
		return err
	}
	fmt.Printf("audit-export: %d entries written to %s\n", n, destPath)
	return nil
}
