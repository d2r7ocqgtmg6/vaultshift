package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var pruneCmd = &cobra.Command{
	Use:   "prune <dest-prefix> <keep-path>...",
	Short: "Remove secrets under dest-prefix not present in keep-paths",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPrune,
}

func init() {
	pruneCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	pruneCmd.Flags().Bool("dry-run", false, "preview deletions without applying them")
	pruneCmd.Flags().String("audit-log", "", "path to audit log file (default stdout)")
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
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

	client, err := vault.New(cfg.Dest.Address, cfg.Dest.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	destPrefix := args[0]
	keepPaths := args[1:]

	pruner, err := vault.NewPruner(client, logger, dryRun)
	if err != nil {
		return err
	}

	result, err := pruner.Prune(destPrefix, keepPaths)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "deleted: %d  skipped: %d  errors: %d\n",
		len(result.Deleted), len(result.Skipped), len(result.Errors))
	if len(result.Errors) > 0 {
		return fmt.Errorf("prune completed with %d error(s)", len(result.Errors))
	}
	return nil
}
