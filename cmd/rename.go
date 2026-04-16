package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var renameCmd = &cobra.Command{
	Use:   "rename <src> <dst>",
	Short: "Rename (move) a secret path within the source namespace",
	Args:  cobra.ExactArgs(2),
	RunE:  runRename,
}

func init() {
	renameCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	renameCmd.Flags().Bool("dry-run", false, "preview rename without making changes")
	RootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	renamer, err := vault.NewRenamer(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("renamer: %w", err)
	}

	if err := renamer.Rename(args[0], args[1]); err != nil {
		return err
	}

	log.Printf("renamed %q -> %q (dryRun=%v)\n", args[0], args[1], dryRun)
	return nil
}
