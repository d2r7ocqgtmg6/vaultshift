package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var namespaceCmd = &cobra.Command{
	Use:   "namespace <src-prefix> <dst-prefix>",
	Short: "Move all secrets from one namespace prefix to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runNamespace,
}

func init() {
	namespaceCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	namespaceCmd.Flags().Bool("dry-run", false, "Preview moves without writing")
	rootCmd.AddCommand(namespaceCmd)
}

func runNamespace(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("namespace: load config: %w", err)
	}

	log, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("namespace: init logger: %w", err)
	}

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("namespace: source client: %w", err)
	}

	dst, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("namespace: dest client: %w", err)
	}

	mover, err := vault.NewNamespaceMover(src, dst, log, dryRun)
	if err != nil {
		return err
	}

	n, err := mover.Move(args[0], args[1])
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would move %d secret(s) from %q to %q\n", n, args[0], args[1])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "moved %d secret(s) from %q to %q\n", n, args[0], args[1])
	}
	return nil
}
