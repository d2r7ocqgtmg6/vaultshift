package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var indexCmd = &cobra.Command{
	Use:   "index <prefix>",
	Short: "Build a key index for all secrets under a prefix",
	Args:  cobra.ExactArgs(1),
	RunE:  runIndex,
}

func init() {
	indexCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	indexCmd.Flags().String("lookup", "", "key name to look up in the index")
	rootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	lookupKey, _ := cmd.Flags().GetString("lookup")
	prefix := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	idxr, err := vault.NewIndexer(client)
	if err != nil {
		return fmt.Errorf("creating indexer: %w", err)
	}

	idx, err := idxr.Build(prefix)
	if err != nil {
		return fmt.Errorf("building index: %w", err)
	}

	if lookupKey != "" {
		paths := idx.Lookup(lookupKey)
		if len(paths) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "no paths found containing key %q\n", lookupKey)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "paths containing key %q:\n", lookupKey)
		for _, p := range paths {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p)
		}
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "index for prefix %q (%d paths):\n", prefix, len(idx))
	for path, entry := range idx {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s  [%s]\n", path, strings.Join(entry.Keys, ", "))
	}
	return nil
}
