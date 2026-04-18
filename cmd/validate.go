package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var validateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a secret against defined rules",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().String("config", ".vaultshift.yaml", "config file")
	validateCmd.Flags().Bool("dry-run", false, "log results without side effects")
	validateCmd.Flags().StringSlice("required-keys", nil, "comma-separated list of required keys")
	validateCmd.Flags().Bool("no-empty-values", false, "fail if any value is empty")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	requiredKeys, _ := cmd.Flags().GetStringSlice("required-keys")
	noEmpty, _ := cmd.Flags().GetBool("no-empty-values")

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

	var rules []vault.ValidationRule
	if len(requiredKeys) > 0 {
		rules = append(rules, vault.RequiredKeys(requiredKeys...))
	}
	if noEmpty {
		rules = append(rules, vault.NoEmptyValues())
	}
	if len(rules) == 0 {
		return fmt.Errorf("specify at least one rule (--required-keys or --no-empty-values)")
	}

	v, err := vault.NewValidator(client, logger, rules, dryRun)
	if err != nil {
		return err
	}

	results, err := v.Validate(args[0])
	if err != nil {
		return err
	}

	allPassed := true
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
			allPassed = false
		}
		fmt.Fprintf(os.Stdout, "[%s] %s\n", status, strings.TrimSpace(r.Reason))
	}
	if !allPassed {
		return fmt.Errorf("one or more validation rules failed")
	}
	return nil
}
