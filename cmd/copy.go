package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var copyCmd = &cobra.Command{
	Use:   "copy <src-path> <dest-path>",
	Short: "Copy a single secret from one path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runCopy,
}

func init() {
	copyCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	copyCmd.Flags().Bool("dry-run", false, "Preview copy without writing")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
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

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}

	dest, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("destination client: %w", err)
	}

	cp, err := vault.NewCopier(src, dest, logger, dryRun)
	if err != nil {
		return err
	}

	if err := cp.Copy(args[0], args[1]); err != nil {
		return err
	}

	log.Printf("copy complete: %s -> %s", args[0], args[1])
	return nil
}
