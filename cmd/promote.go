package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultshift/vaultshift/internal/audit"
	"github.com/vaultshift/vaultshift/internal/config"
	"github.com/vaultshift/vaultshift/internal/vault"
)

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote secrets from a staging prefix to a production prefix",
	RunE:  runPromote,
}

func init() {
	promoteCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	promoteCmd.Flags().String("src-prefix", "", "source prefix to promote from (required)")
	promoteCmd.Flags().String("dst-prefix", "", "destination prefix to promote to (required)")
	promoteCmd.Flags().Bool("skip-existing", false, "skip secrets that already exist in destination")
	promoteCmd.Flags().Bool("dry-run", false, "preview promotion without writing")
	_ = promoteCmd.MarkFlagRequired("src-prefix")
	_ = promoteCmd.MarkFlagRequired("dst-prefix")
	rootCmd.AddCommand(promoteCmd)
}

func runPromote(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	srcPrefix, _ := cmd.Flags().GetString("src-prefix")
	dstPrefix, _ := cmd.Flags().GetString("dst-prefix")
	skipExisting, _ := cmd.Flags().GetBool("skip-existing")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	srcClient, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}
	dstClient, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("destination client: %w", err)
	}

	promoter := vault.NewPromoter(srcClient, dstClient, logger, dryRun)
	results, err := promoter.Promote(srcPrefix, dstPrefix, skipExisting)
	if err != nil {
		return err
	}

	var errCount int
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(os.Stderr, "ERROR  %s: %v\n", r.Path, r.Err)
			errCount++
		case r.Skipped:
			fmt.Printf("SKIP   %s\n", r.Path)
		case r.DryRun:
			fmt.Printf("DRY    %s\n", r.Path)
		default:
			fmt.Printf("OK     %s\n", r.Path)
		}
	}
	if errCount > 0 {
		return fmt.Errorf("%d secret(s) failed to promote", errCount)
	}
	return nil
}
