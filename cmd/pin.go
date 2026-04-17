package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var pinCmd = &cobra.Command{
	Use:   "pin <path>",
	Short: "Pin a secret to prevent accidental overwrite or deletion",
	Args:  cobra.ExactArgs(1),
	RunE:  runPin,
}

func init() {
	pinCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	pinCmd.Flags().Bool("dry-run", false, "Preview pin without writing")
	pinCmd.Flags().String("audit-log", "", "Path to audit log file")
	rootCmd.AddCommand(pinCmd)
}

func runPin(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("init audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("init vault client: %w", err)
	}

	pinner, err := vault.NewPinner(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("init pinner: %w", err)
	}

	result := pinner.Pin(args[0])
	if result.Error != nil {
		return fmt.Errorf("pin failed: %w", result.Error)
	}
	if result.Skipped {
		fmt.Println("already pinned:", args[0])
	} else if dryRun {
		fmt.Println("[dry-run] would pin:", args[0])
	} else {
		fmt.Println("pinned:", args[0])
	}
	return nil
}
