package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var replicateCmd = &cobra.Command{
	Use:   "replicate <src-prefix> <dest-prefix>[,<dest-prefix>...]",
	Short: "Replicate secrets from one prefix to multiple destinations",
	Args:  cobra.ExactArgs(2),
	RunE:  runReplicate,
}

func init() {
	replicateCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	replicateCmd.Flags().Bool("dry-run", false, "preview replication without writing")
	replicateCmd.Flags().Bool("overwrite", false, "overwrite existing secrets at destination")
	replicateCmd.Flags().String("audit-log", "", "path to audit log file (default stdout)")
	rootCmd.AddCommand(replicateCmd)
}

func runReplicate(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	srcPrefix := args[0]
	destPrefixes := strings.Split(args[1], ",")

	replicator, err := vault.NewReplicator(client, logger, dryRun, overwrite)
	if err != nil {
		return err
	}

	results, err := replicator.Replicate(srcPrefix, destPrefixes)
	if err != nil {
		return err
	}

	for _, res := range results {
		switch {
		case res.Error != nil:
			fmt.Fprintf(cmd.ErrOrStderr(), "ERROR  %s: %v\n", res.Path, res.Error)
		case res.Skipped:
			fmt.Fprintf(cmd.OutOrStdout(), "SKIP   %s\n", res.Path)
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "OK     %s\n", res.Path)
		}
	}
	return nil
}
