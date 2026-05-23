package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror <src-prefix> <dst-prefix>",
	Short: "Mirror all secrets from a source prefix to a destination prefix",
	Args:  cobra.ExactArgs(2),
	RunE:  runMirror,
}

func init() {
	mirrorCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "path to config file")
	mirrorCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	mirrorCmd.Flags().StringP("audit-log", "a", "", "path to audit log file (defaults to stdout)")
	rootCmd.AddCommand(mirrorCmd)
}

func runMirror(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("mirror: load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("mirror: init logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("mirror: init client: %w", err)
	}

	mirrorer, err := vault.NewMirrorer(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("mirror: %w", err)
	}

	results, err := mirrorer.Mirror(args[0], args[1])
	if err != nil {
		return fmt.Errorf("mirror: %w", err)
	}

	for _, r := range results {
		switch {
		case r.Error != nil:
			fmt.Fprintf(os.Stderr, "ERROR  %s: %v\n", r.Path, r.Error)
		case r.Skipped:
			fmt.Printf("DRY-RUN  %s\n", r.Path)
		default:
			fmt.Printf("MIRRORED %s\n", r.Path)
		}
	}
	return nil
}
