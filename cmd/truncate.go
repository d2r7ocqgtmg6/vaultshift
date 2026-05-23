package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

func init() {
	truncateCmd := &cobra.Command{
		Use:   "truncate <path>",
		Short: "Truncate secret values at a path to a maximum length",
		Args:  cobra.ExactArgs(1),
		RunE:  runTruncate,
	}
	truncateCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	truncateCmd.Flags().Int("max-len", 64, "maximum character length for secret values")
	truncateCmd.Flags().StringSlice("keys", nil, "comma-separated list of keys to truncate")
	truncateCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	truncateCmd.Flags().String("audit-log", "", "path to audit log file (default stdout)")
	rootCmd.AddCommand(truncateCmd)
}

func runTruncate(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	maxLen, _ := cmd.Flags().GetInt("max-len")
	keys, _ := cmd.Flags().GetStringSlice("keys")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	if len(keys) == 0 {
		return fmt.Errorf("at least one --keys value is required")
	}

	// strip any whitespace from keys
	for i, k := range keys {
		keys[i] = strings.TrimSpace(k)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.SourceAddress, cfg.SourceToken)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	tr, err := vault.NewTruncater(client, logger, maxLen, keys, dryRun)
	if err != nil {
		return fmt.Errorf("truncater: %w", err)
	}

	path := args[0]
	if err := tr.Truncate(path); err != nil {
		return fmt.Errorf("truncate %s: %w", path, err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] truncate preview logged for %s\n", path)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "truncated values at %s (maxLen=%d)\n", path, maxLen)
	}
	return nil
}
