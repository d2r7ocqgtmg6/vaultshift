package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var observeCmd = &cobra.Command{
	Use:   "observe [paths...]",
	Short: "Observe secrets at given paths and report their status",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runObserve,
}

func init() {
	observeCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file path")
	observeCmd.Flags().Bool("dry-run", false, "print results without side effects")
	rootCmd.AddCommand(observeCmd)
}

func runObserve(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("observe: config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("observe: client: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("observe: logger: %w", err)
	}

	obs, err := vault.NewObserver(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("observe: %w", err)
	}

	results := obs.Observe(cmd.Context(), args)
	for _, r := range results {
		if r.Error != nil {
			fmt.Fprintf(os.Stderr, "ERROR  %s: %v\n", r.Path, r.Error)
			continue
		}
		if !r.Exists {
			fmt.Printf("MISSING %s\n", r.Path)
		} else {
			fmt.Printf("FOUND   %s (keys=%d)\n", r.Path, r.KeyCount)
		}
	}
	return nil
}
