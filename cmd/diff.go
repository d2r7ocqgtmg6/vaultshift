package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare secrets between source and destination namespaces",
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	diffCmd.Flags().String("prefix", "", "Secret prefix to diff (required)")
	_ = diffCmd.MarkFlagRequired("prefix")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	prefix, _ := cmd.Flags().GetString("prefix")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	srcClient, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("creating source client: %w", err)
	}

	dstClient, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("creating destination client: %w", err)
	}

	differ := vault.NewDiffer(srcClient, dstClient)
	results, err := differ.DiffAll(prefix)
	if err != nil {
		return fmt.Errorf("running diff: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No secrets found under prefix:", prefix)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATUS\tKEY")
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\n", r.Status, r.Key)
	}
	w.Flush()
	return nil
}
