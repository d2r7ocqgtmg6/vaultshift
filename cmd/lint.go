package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var lintCmd = &cobra.Command{
	Use:   "lint <path>",
	Short: "Lint a secret against configurable rules",
	Args:  cobra.ExactArgs(1),
	RunE:  runLint,
}

func init() {
	lintCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	lintCmd.Flags().Bool("no-empty", true, "flag keys with empty string values")
	lintCmd.Flags().Bool("no-uppercase", false, "flag keys containing uppercase letters")
	rootCmd.AddCommand(lintCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
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

	noEmpty, _ := cmd.Flags().GetBool("no-empty")
	noUpper, _ := cmd.Flags().GetBool("no-uppercase")

	var rules []vault.LintRule
	if noEmpty {
		rules = append(rules, vault.NoEmptyKeys)
	}
	if noUpper {
		rules = append(rules, vault.NoUpperCaseKeys)
	}
	if len(rules) == 0 {
		rules = append(rules, vault.NoEmptyKeys)
	}

	linter, err := vault.NewLinter(client, logger, rules...)
	if err != nil {
		return fmt.Errorf("linter: %w", err)
	}

	result, err := linter.Lint(args[0])
	if err != nil {
		return err
	}

	if len(result.Violations) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "OK: %s\n", result.Path)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "VIOLATIONS: %s\n", result.Path)
	for _, v := range result.Violations {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", v)
	}
	return fmt.Errorf("lint: %d violation(s) found", len(result.Violations))
}
