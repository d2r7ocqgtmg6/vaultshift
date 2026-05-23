package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var envelopeCmd = &cobra.Command{
	Use:   "envelope <path> [path...]",
	Short: "Wrap secrets into a portable envelope file",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runEnvelope,
}

func init() {
	envelopeCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	envelopeCmd.Flags().String("out", "envelope.json", "output file path")
	envelopeCmd.Flags().Bool("dry-run", false, "preview without writing the file")
	rootCmd.AddCommand(envelopeCmd)
}

func runEnvelope(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	outPath, _ := cmd.Flags().GetString("out")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("envelope: load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("envelope: audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("envelope: vault client: %w", err)
	}

	enveloper, err := vault.NewEnveloper(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("envelope: %w", err)
	}

	env, err := enveloper.Wrap(args, outPath)
	if err != nil {
		return fmt.Errorf("envelope: wrap: %w", err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would write %d secret(s) to %s\n", len(env.Entries), outPath)
		for _, e := range env.Entries {
			keys := make([]string, 0, len(e.Data))
			for k := range e.Data {
				keys = append(keys, k)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s [%s]\n", e.Path, strings.Join(keys, ", "))
		}
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "wrote %d secret(s) to %s\n", len(env.Entries), outPath)
	return nil
}
