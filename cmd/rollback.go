package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Undo the most recent migration by deleting written secrets",
	Long: `rollback deletes every secret that was written to the destination
Vault during the last migrate run. It relies on the audit log to
determine which paths to remove.`,
	RunE: runRollback,
}

func init() {
	rollbackCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	rollbackCmd.Flags().String("audit-log", "", "path to audit log (defaults to stdout)")
	rollbackCmd.Flags().Bool("dry-run", false, "print paths that would be deleted without deleting them")
	RootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	auditPath, _ := cmd.Flags().GetString("audit-log")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	l, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	dstClient, err := vault.New(cfg.Dest.Address, cfg.Dest.Token)
	if err != nil {
		return fmt.Errorf("dest vault client: %w", err)
	}

	rb := vault.NewRollbacker(dstClient, l)

	// Populate rollback records from config-defined paths.
	for _, p := range cfg.Paths {
		rb.Record(cfg.Dest.Mount, p.Dest)
	}

	if rb.Len() == 0 {
		fmt.Fprintln(os.Stdout, "no paths configured for rollback")
		return nil
	}

	if dryRun {
		return printDryRun(rb.Paths())
	}

	l.Log("rollback_start", map[string]interface{}{"count": rb.Len()})
	errs := rb.Rollback(cmd.Context())
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "rollback error: %v\n", e)
		}
		return fmt.Errorf("%d rollback error(s) encountered", len(errs))
	}

	fmt.Fprintf(os.Stdout, "rollback complete: %d secret(s) deleted\n", rb.Len())
	return nil
}

// printDryRun lists the paths that would be deleted without performing any
// deletions. Useful for verifying rollback scope before committing.
func printDryRun(paths []string) error {
	fmt.Fprintf(os.Stdout, "dry-run: %d path(s) would be deleted:\n", len(paths))
	for _, p := range paths {
		fmt.Fprintf(os.Stdout, "  - %s\n", p)
	}
	return nil
}
