package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var splitCmd = &cobra.Command{
	Use:   "split <source-path>",
	Short: "Split a secret's keys into multiple destination paths based on routing rules",
	Args:  cobra.ExactArgs(1),
	RunE:  runSplit,
}

func init() {
	splitCmd.Flags().StringToString("routes", nil, "key=dest-prefix routing map (e.g. db_pass=secret/db)")
	splitCmd.Flags().Bool("dry-run", false, "Preview splits without writing")
	splitCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	rootCmd.AddCommand(splitCmd)
}

func runSplit(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("split: load config: %w", err)
	}

	routes, _ := cmd.Flags().GetStringToString("routes")
	if len(routes) == 0 {
		return fmt.Errorf("split: --routes is required")
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("split: vault client: %w", err)
	}

	logger, err := newAuditLogger(cfg)
	if err != nil {
		return fmt.Errorf("split: audit logger: %w", err)
	}

	splitter, err := vault.NewSplitter(client, logger, routes, dryRun)
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}

	results, err := splitter.Split(args[0])
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}

	for _, r := range results {
		status := "written"
		if r.Skipped {
			status = "dry-run"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s [%s]\n", r.SourcePath, r.DestPath, status)
	}
	return nil
}
