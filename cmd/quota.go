package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultshift/internal/audit"
	"github.com/your-org/vaultshift/internal/config"
	"github.com/your-org/vaultshift/internal/vault"
)

var quotaCmd = &cobra.Command{
	Use:   "quota [prefix...]",
	Short: "Check secret count quotas under one or more prefixes",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runQuota,
}

func init() {
	quotaCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	quotaCmd.Flags().String("audit-log", "", "path to audit log file (default stdout)")
	quotaCmd.Flags().Int("limit", 100, "maximum number of secrets allowed per prefix")
	rootCmd.AddCommand(quotaCmd)
}

func runQuota(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	auditPath, _ := cmd.Flags().GetString("audit-log")
	limit, _ := cmd.Flags().GetInt("limit")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("creating audit logger: %w", err)
	}

	quoter, err := vault.NewQuoter(client, logger, limit)
	if err != nil {
		return fmt.Errorf("creating quoter: %w", err)
	}

	results, err := quoter.Check(args)
	if err != nil {
		return fmt.Errorf("quota check: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PREFIX\tCOUNT\tLIMIT\tSTATUS")
	for _, r := range results {
		status := "OK"
		if r.Exceeds {
			status = "EXCEEDED"
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", r.Path, r.Count, limit, status)
	}
	return w.Flush()
}
