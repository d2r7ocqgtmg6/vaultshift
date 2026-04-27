package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint <prefix> <output-file>",
	Short: "Save a checkpoint of secrets under a prefix to a local file",
	Long: `checkpoint reads all secrets under the given Vault prefix and writes
them to a local file that can later be restored with the 'restore' command.

Use --dry-run to preview the operation without writing any data to disk.`,
	Args: cobra.ExactArgs(2),
	RunE: runCheckpoint,
}

func init() {
	checkpointCmd.Flags().String("config", ".vaultshift.yaml", "config file path")
	checkpointCmd.Flags().Bool("dry-run", false, "preview without writing")
	rootCmd.AddCommand(checkpointCmd)
}

func runCheckpoint(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	prefix := args[0]
	outPath := args[1]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	cp, err := vault.NewCheckpointer(client)
	if err != nil {
		return fmt.Errorf("checkpointer: %w", err)
	}

	if err := cp.Save(prefix, outPath, dryRun); err != nil {
		return fmt.Errorf("checkpoint: %w", err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would save checkpoint of %s to %s\n", prefix, outPath)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "checkpoint saved: %s\n", outPath)
	}
	return nil
}
