package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

func init() {
	alignCmd := &cobra.Command{
		Use:   "align <prefix>",
		Short: "Align destination secrets to match source under a given prefix",
		Args:  cobra.ExactArgs(1),
		RunE:  runAlign,
	}
	alignCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	alignCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	alignCmd.Flags().Bool("prune", false, "Remove destination secrets absent from source")
	alignCmd.Flags().String("audit-log", "", "Path to audit log file (default stdout)")
	rootCmd.AddCommand(alignCmd)
}

func runAlign(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	prune, _ := cmd.Flags().GetBool("prune")
	auditPath, _ := cmd.Flags().GetString("audit-log")
	prefix := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token, cfg.Source.Namespace)
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}
	dst, err := vault.New(cfg.Dest.Address, cfg.Dest.Token, cfg.Dest.Namespace)
	if err != nil {
		return fmt.Errorf("dest client: %w", err)
	}

	aligner, err := vault.NewAligner(src, dst, logger, dryRun, prune)
	if err != nil {
		return fmt.Errorf("aligner: %w", err)
	}

	result, err := aligner.Align(prefix)
	if err != nil {
		return fmt.Errorf("align: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "written: %d  pruned: %d  errors: %d\n",
		len(result.Written), len(result.Pruned), len(result.Errors))
	for _, e := range result.Errors {
		fmt.Fprintf(cmd.ErrOrStderr(), "error: %s\n", e)
	}
	return nil
}
