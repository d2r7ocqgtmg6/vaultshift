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
	evictCmd := &cobra.Command{
		Use:   "evict <path> <pattern,...>",
		Short: "Remove secret keys matching patterns at a given path",
		Args:  cobra.ExactArgs(2),
		RunE:  runEvict,
	}
	evictCmd.Flags().Bool("dry-run", false, "Preview evictions without writing changes")
	evictCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	evictCmd.Flags().String("audit-log", "", "Path to audit log file (default: stdout)")
	rootCmd.AddCommand(evictCmd)
}

func runEvict(cmd *cobra.Command, args []string) error {
	path := args[0]
	patterns := strings.Split(args[1], ",")

	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	e, err := vault.NewEvicter(client, logger, patterns, dryRun)
	if err != nil {
		return fmt.Errorf("evicter: %w", err)
	}

	results, err := e.Evict(path)
	if err != nil {
		return err
	}

	for _, r := range results {
		if r.Evicted {
			status := "evicted"
			if dryRun {
				status = "would evict"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s [%s]\n", r.Path, r.Key, status)
		}
	}
	return nil
}
