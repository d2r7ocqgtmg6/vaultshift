package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var packCmd = &cobra.Command{
	Use:   "pack <path>... --out <file>",
	Short: "Bundle multiple secrets into a single JSON pack file",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPack,
}

func init() {
	packCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	packCmd.Flags().String("out", "secrets.pack.json", "output pack file")
	packCmd.Flags().Bool("dry-run", false, "preview without writing")
	rootCmd.AddCommand(packCmd)
}

func runPack(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	outFile, _ := cmd.Flags().GetString("out")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	packer, err := vault.NewPacker(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("packer: %w", err)
	}

	if err := packer.Pack(args, outFile); err != nil {
		return fmt.Errorf("pack: %w", err)
	}

	if dryRun {
		fmt.Printf("[dry-run] would pack %d path(s) to %s\n", len(args), outFile)
	} else {
		fmt.Printf("packed %d path(s) → %s\n", len(args), strings.Join(args, ", "))
	}
	return nil
}
