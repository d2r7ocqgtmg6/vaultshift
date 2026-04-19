package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var inheritCmd = &cobra.Command{
	Use:   "inherit <parent-path> <child-path>",
	Short: "Inherit secrets from a parent path into a child path",
	Args:  cobra.ExactArgs(2),
	RunE:  runInherit,
}

func init() {
	inheritCmd.Flags().String("config", ".vaultshift.yaml", "config file")
	inheritCmd.Flags().Bool("dry-run", false, "preview without writing")
	inheritCmd.Flags().Bool("overwrite", false, "overwrite existing keys in child")
	rootCmd.AddCommand(inheritCmd)
}

func runInherit(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	overwrite, _ := cmd.Flags().GetBool("overwrite")

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

	h, err := vault.NewInheritor(client, logger, dryRun, overwrite)
	if err != nil {
		return err
	}

	n, err := h.Inherit(args[0], args[1])
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("[dry-run] would inherit %d key(s) from %q into %q\n", n, args[0], args[1])
	} else {
		fmt.Printf("inherited %d key(s) from %q into %q\n", n, args[0], args[1])
	}
	return nil
}
