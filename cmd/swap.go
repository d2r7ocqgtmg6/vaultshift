package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var swapCmd = &cobra.Command{
	Use:   "swap <path-a> <path-b>",
	Short: "Exchange the values of two secrets",
	Args:  cobra.ExactArgs(2),
	RunE:  runSwap,
}

func init() {
	swapCmd.Flags().String("config", ".vaultshift.yaml", "config file path")
	swapCmd.Flags().Bool("dry-run", false, "preview swap without writing")
	swapCmd.Flags().String("audit-log", "", "path to audit log file (default stdout)")
	rootCmd.AddCommand(swapCmd)
}

func runSwap(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("swap: load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("swap: audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token, cfg.Source.Namespace)
	if err != nil {
		return fmt.Errorf("swap: vault client: %w", err)
	}

	sw, err := vault.NewSwapper(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("swap: %w", err)
	}

	if err := sw.Swap(args[0], args[1]); err != nil {
		return fmt.Errorf("swap: %w", err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would swap %s <-> %s\n", args[0], args[1])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "swapped %s <-> %s\n", args[0], args[1])
	}
	return nil
}
