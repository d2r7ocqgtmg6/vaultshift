package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var stashCmd = &cobra.Command{
	Use:   "stash [paths...]",
	Short: "Temporarily save secrets to a local JSONL file",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runStash,
}

func init() {
	stashCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	stashCmd.Flags().String("out", "stash.jsonl", "output JSONL file for stashed secrets")
	stashCmd.Flags().Bool("dry-run", false, "preview without writing file")
	rootCmd.AddCommand(stashCmd)
}

func runStash(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	out, _ := cmd.Flags().GetString("out")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(vault.Config{
		Address: cfg.Source.Address,
		Token:   cfg.Source.Token,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	stasher, err := vault.NewStasher(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("stasher: %w", err)
	}

	entries, err := stasher.Stash(args, out)
	if err != nil {
		return err
	}

	paths := make([]string, len(entries))
	for i, e := range entries {
		paths[i] = e.Path
	}

	if dryRun {
		fmt.Printf("[dry-run] would stash %d secret(s): %s\n", len(entries), strings.Join(paths, ", "))
	} else {
		fmt.Printf("stashed %d secret(s) to %s\n", len(entries), out)
	}
	return nil
}
