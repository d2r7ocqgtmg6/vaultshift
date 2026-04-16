package cmd

import (
	"fmt"
	"log"

	"github.com/drew/vaultshift/internal/audit"
	"github.com/drew/vaultshift/internal/config"
	"github.com/drew/vaultshift/internal/vault"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [prefix] [dest-file]",
	Short: "Export secrets from a Vault prefix to a local JSON file",
	Args:  cobra.ExactArgs(2),
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	exportCmd.Flags().Bool("dry-run", false, "Preview export without writing file")
	exportCmd.Flags().String("audit-log", "", "Path to audit log file")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token, cfg.Source.Mount)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	exporter, err := vault.NewExporter(client, logger)
	if err != nil {
		return err
	}

	prefix := args[0]
	dest := args[1]

	n, err := exporter.Export(prefix, dest, dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		log.Printf("[dry-run] would export %d secrets from %s", n, prefix)
	} else {
		log.Printf("exported %d secrets to %s", n, dest)
	}
	return nil
}
