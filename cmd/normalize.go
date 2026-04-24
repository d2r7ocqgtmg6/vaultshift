package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var normalizeCmd = &cobra.Command{
	Use:   "normalize <path>",
	Short: "Normalize secret keys and values at the given path",
	Args:  cobra.ExactArgs(1),
	RunE:  runNormalize,
}

func init() {
	normalizeCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "path to config file")
	normalizeCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	normalizeCmd.Flags().Bool("trim-space", true, "trim whitespace from string values")
	normalizeCmd.Flags().Bool("lower-keys", false, "lowercase all secret keys")
	rootCmd.AddCommand(normalizeCmd)
}

func runNormalize(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	trimSpace, _ := cmd.Flags().GetBool("trim-space")
	lowerKeys, _ := cmd.Flags().GetBool("lower-keys")

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

	opts := []vault.NormalizerOption{}
	if dryRun {
		opts = append(opts, vault.WithNormalizeDryRun())
	}
	if trimSpace {
		opts = append(opts, vault.WithNormalizeTrimSpace())
	}
	if lowerKeys {
		opts = append(opts, vault.WithNormalizeLowerKeys())
	}

	n, err := vault.NewNormalizer(client, logger, opts...)
	if err != nil {
		return fmt.Errorf("normalizer: %w", err)
	}

	path := args[0]
	if err := n.Normalize(path); err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would normalize %s\n", path)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "normalized %s\n", path)
	}
	return nil
}
