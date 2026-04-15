package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var expireCmd = &cobra.Command{
	Use:   "expire",
	Short: "Purge secrets whose expiry timestamp has passed",
	RunE:  runExpire,
}

func init() {
	expireCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	expireCmd.Flags().String("meta-key", "expires_at", "Secret data key holding RFC3339 expiry timestamp")
	expireCmd.Flags().String("prefix", "", "Vault path prefix to scan for expiring secrets")
	expireCmd.Flags().Bool("dry-run", false, "Preview expirations without deleting")
	expireCmd.Flags().String("audit-log", "", "Path to audit log file (stdout if empty)")
	rootCmd.AddCommand(expireCmd)
}

func runExpire(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	metaKey, _ := cmd.Flags().GetString("meta-key")
	prefix, _ := cmd.Flags().GetString("prefix")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	lister := vault.NewLister(client)
	paths, err := lister.List(prefix)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}

	expirer := vault.NewExpirer(client, logger, metaKey, dryRun)
	results := expirer.CheckAndPurge(paths)

	var errs []string
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.Path, r.Error))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("expire errors:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}
