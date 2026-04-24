package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultshift/internal/audit"
	"github.com/your-org/vaultshift/internal/config"
	"github.com/your-org/vaultshift/internal/vault"
)

func init() {
	flattenCmd := &cobra.Command{
		Use:   "flatten <src-prefix> <dst-path>",
		Short: "Merge all secrets under a prefix into a single destination secret",
		Args:  cobra.ExactArgs(2),
		RunE:  runFlatten,
	}

	flattenCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "path to config file")
	flattenCmd.Flags().Bool("dry-run", false, "preview merge without writing")
	flattenCmd.Flags().String("audit-log", "", "path to audit log file (default stdout)")

	rootCmd.AddCommand(flattenCmd)
}

func runFlatten(cmd *cobra.Command, args []string) error {
	srcPrefix := args[0]
	dstPath := args[1]

	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("flatten: load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("flatten: audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("flatten: vault client: %w", err)
	}

	flattener, err := vault.NewFlattener(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("flatten: init: %w", err)
	}

	if err := flattener.Flatten(srcPrefix, dstPath); err != nil {
		return fmt.Errorf("flatten: %w", err)
	}

	if dryRun {
		cmd.Println("[dry-run] flatten complete — no secrets written")
	} else {
		cmd.Printf("Flattened %q → %q\n", srcPrefix, dstPath)
	}
	return nil
}
