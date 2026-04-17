package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var redactCmd = &cobra.Command{
	Use:   "redact <path>",
	Short: "Redact specific keys in a secret by replacing their values with a placeholder",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedact,
}

func init() {
	redactCmd.Flags().String("config", ".vaultshift.yaml", "config file")
	redactCmd.Flags().StringSlice("keys", nil, "comma-separated list of keys to redact (required)")
	redactCmd.Flags().String("placeholder", "REDACTED", "value to substitute for redacted keys")
	redactCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	_ = redactCmd.MarkFlagRequired("keys")
	rootCmd.AddCommand(redactCmd)
}

func runRedact(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	keys, _ := cmd.Flags().GetStringSlice("keys")
	placeholder, _ := cmd.Flags().GetString("placeholder")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	path := args[0]

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	redactor, err := vault.NewRedactor(client, logger, keys, placeholder)
	if err != nil {
		return fmt.Errorf("redactor: %w", err)
	}

	if err := redactor.Redact(path, dryRun); err != nil {
		return fmt.Errorf("redact: %w", err)
	}

	if dryRun {
		fmt.Printf("[dry-run] would redact keys [%s] at %s\n", strings.Join(keys, ", "), path)
	} else {
		fmt.Printf("redacted keys [%s] at %s\n", strings.Join(keys, ", "), path)
	}
	return nil
}
