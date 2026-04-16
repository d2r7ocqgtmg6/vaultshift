package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <path>",
	Short: "Archive a secret to a timestamped path and remove the original",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchive,
}

func init() {
	archiveCmd.Flags().String("config", ".vaultshift.yaml", "config file")
	archiveCmd.Flags().String("archive-prefix", "archive", "prefix for archived secrets")
	archiveCmd.Flags().Bool("dry-run", false, "preview without making changes")
	archiveCmd.Flags().String("audit-log", "", "path to audit log file")
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	archivePrefix, _ := cmd.Flags().GetString("archive-prefix")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	archiver, err := vault.NewArchiver(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("archiver: %w", err)
	}

	if err := archiver.Archive(args[0], archivePrefix); err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	fmt.Printf("archived %s -> %s/\n", args[0], archivePrefix)
	return nil
}
