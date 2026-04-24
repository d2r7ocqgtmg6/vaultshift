package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultshift/internal/audit"
	"github.com/your-org/vaultshift/internal/config"
	"github.com/your-org/vaultshift/internal/vault"
)

var auditSummaryCmd = &cobra.Command{
	Use:   "audit-summary <audit-log-file>",
	Short: "Summarize an audit log file produced by vaultshift",
	Args:  cobra.ExactArgs(1),
	RunE:  runAuditSummary,
}

func init() {
	auditSummaryCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "path to config file")
	auditSummaryCmd.Flags().StringP("audit-log", "a", "", "path to audit log output (stdout if empty)")
	rootCmd.AddCommand(auditSummaryCmd)
}

func runAuditSummary(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	_, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	auditPath, _ := cmd.Flags().GetString("audit-log")
	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("create audit logger: %w", err)
	}

	summarizer, err := vault.NewAuditSummarizer(logger)
	if err != nil {
		return fmt.Errorf("create summarizer: %w", err)
	}

	f, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("open audit log %q: %w", args[0], err)
	}
	defer f.Close()

	summary, err := summarizer.Summarize(f)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	summarizer.Print(cmd.OutOrStdout(), summary)
	return nil
}
