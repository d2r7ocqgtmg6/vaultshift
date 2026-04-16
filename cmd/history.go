package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var historyCmd = &cobra.Command{
	Use:   "history <secret-path>",
	Short: "List version history of a secret in the source Vault",
	Args:  cobra.ExactArgs(1),
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	historyCmd.Flags().String("mount", "secret", "KV v2 mount path")
	historyCmd.Flags().Bool("data", false, "fetch and display data for each version")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	mount, _ := cmd.Flags().GetString("mount")
	showData, _ := cmd.Flags().GetBool("data")
	secretPath := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	h, err := vault.NewHistorian(client, mount)
	if err != nil {
		return fmt.Errorf("historian: %w", err)
	}

	ctx := context.Background()
	entries, err := h.ListVersions(ctx, secretPath)
	if err != nil {
		return fmt.Errorf("list versions: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tCREATED\tDELETED")
	for _, e := range entries {
		del := "-"
		if e.Deleted {
			del = "yes"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\n", e.Version, e.CreatedAt.Format("2006-01-02T15:04:05Z"), del)
		if showData {
			ve, err := h.ReadVersion(ctx, secretPath, e.Version)
			if err != nil {
				fmt.Fprintf(w, "  (error reading data: %v)\n", err)
				continue
			}
			for k, v := range ve.Data {
				fmt.Fprintf(w, "  %s\t=\t%v\n", k, v)
			}
		}
	}
	return w.Flush()
}
