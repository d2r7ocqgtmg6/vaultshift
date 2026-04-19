package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var policyCmd = &cobra.Command{
	Use:   "policy [paths...]",
	Short: "Check Vault policy capabilities for given paths",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPolicy,
}

func init() {
	policyCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file")
	policyCmd.Flags().StringP("audit-log", "a", "", "audit log file path")
	rootCmd.AddCommand(policyCmd)
}

func runPolicy(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	checker, err := vault.NewPolicyChecker(client, logger)
	if err != nil {
		return fmt.Errorf("failed to create policy checker: %w", err)
	}

	results, err := checker.Check(args)
	if err != nil {
		return fmt.Errorf("policy check failed: %w", err)
	}

	for _, r := range results {
		status := "DENIED"
		if r.Allowed {
			status = "ALLOWED"
		}
		caps := strings.Join(r.Capabilities, ", ")
		fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s (%s)\n", status, r.Path, caps)
	}
	return nil
}
