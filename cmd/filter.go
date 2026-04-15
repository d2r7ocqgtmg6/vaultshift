package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultshift/internal/config"
	"github.com/yourorg/vaultshift/internal/vault"
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	 Short: "Preview which secret paths would be included or excluded by current filter rules",
	RunE:  runFilter,
}

func init() {
	rootCmd.AddCommand(filterCmd)
	filterCmd.Flags().StringP("config", "c", "", "Path to config file")
	filterCmd.Flags().String("prefix", "", "Root prefix to list secrets from")
}

func runFilter(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	prefix, _ := cmd.Flags().GetString("prefix")
	if prefix == "" {
		prefix = cfg.Prefix
	}

	src, err := vault.New(cfg.SourceAddress, cfg.SourceToken, cfg.SourceNamespace)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	paths, err := vault.ListSecrets(cmd.Context(), src, prefix)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}

	f := vault.NewFilter(cfg.IncludePrefixes, cfg.ExcludePrefixes)
	allowed := f.FilterPaths(paths)

	fmt.Fprintf(cmd.OutOrStdout(), "Total paths found : %d\n", len(paths))
	fmt.Fprintf(cmd.OutOrStdout(), "Paths after filter: %d\n", len(allowed))
	for _, p := range allowed {
		fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", p)
	}
	return nil
}
