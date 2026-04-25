package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var graphCmd = &cobra.Command{
	Use:   "graph <prefix>",
	Short: "Render a dependency graph of secret paths under a prefix",
	Args:  cobra.ExactArgs(1),
	RunE:  runGraph,
}

func init() {
	graphCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "path to config file")
	graphCmd.Flags().StringP("output", "o", "text", "output format: text or dot")
	rootCmd.AddCommand(graphCmd)
}

func runGraph(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	output, _ := cmd.Flags().GetString("output")
	prefix := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	grapher, err := vault.NewGrapher(client, logger)
	if err != nil {
		return fmt.Errorf("grapher: %w", err)
	}

	result, err := grapher.Build(prefix)
	if err != nil {
		return fmt.Errorf("build graph: %w", err)
	}

	switch output {
	case "dot":
		fmt.Fprintln(cmd.OutOrStdout(), "digraph secrets {")
		for _, edge := range result.Edges {
			fmt.Fprintf(cmd.OutOrStdout(), "  %q -> %q;\n", edge[0], edge[1])
		}
		fmt.Fprintln(cmd.OutOrStdout(), "}")
	default:
		for _, node := range result.Nodes {
			indent := strings.Repeat("  ", node.Depth)
			fmt.Fprintf(cmd.OutOrStdout(), "%s%s (%d children)\n", indent, node.Path, len(node.Children))
		}
	}

	return nil
}
