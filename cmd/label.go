package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var labelCmd = &cobra.Command{
	Use:   "label [paths...]",
	Short: "Attach or remove labels on secrets",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runLabel,
}

func init() {
	labelCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file")
	labelCmd.Flags().StringSliceP("set", "s", nil, "labels to set in key=value format")
	labelCmd.Flags().StringSliceP("remove", "r", nil, "label keys to remove")
	labelCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	rootCmd.AddCommand(labelCmd)
}

func runLabel(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	setRaw, _ := cmd.Flags().GetStringSlice("set")
	remove, _ := cmd.Flags().GetStringSlice("remove")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	labels := make(map[string]string, len(setRaw))
	for _, kv := range setRaw {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label %q: expected key=value", kv)
		}
		labels[parts[0]] = parts[1]
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	labeler, err := vault.NewLabeler(client, logger, labels, remove, dryRun)
	if err != nil {
		return err
	}

	if err := labeler.Label(args); err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "dry-run: no changes written")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "labeled %d path(s)\n", len(args))
	}
	return nil
}
