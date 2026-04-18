package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/subtlepseudonym/vaultshift/internal/audit"
	"github.com/subtlepseudonym/vaultshift/internal/config"
	"github.com/subtlepseudonym/vaultshift/internal/vault"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <path>",
	Short: "Rotate secret values at the given path",
	Args:  cobra.ExactArgs(1),
	RunE:  runRotate,
}

func init() {
	rotateCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	rotateCmd.Flags().Bool("dry-run", false, "preview rotation without writing")
	rotateCmd.Flags().String("audit-log", "", "path to audit log file")
	rootCmd.AddCommand(rotateCmd)
}

func runRotate(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("rotate: load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("rotate: audit logger: %w", err)
	}

	client, err := vault.New(vault.Config{
		Address: cfg.Source.Address,
		Token:   cfg.Source.Token,
	})
	if err != nil {
		return fmt.Errorf("rotate: vault client: %w", err)
	}

	rotator, err := vault.NewRotator(client, logger, vault.WithDryRun(dryRun))
	if err != nil {
		return fmt.Errorf("rotate: %w", err)
	}

	result, err := rotator.Rotate(args[0])
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would rotate %d key(s) at %s\n", len(result), args[0])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "rotated %d key(s) at %s\n", len(result), args[0])
	}
	return nil
}
