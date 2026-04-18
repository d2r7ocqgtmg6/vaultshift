package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var trimCmd = &cobra.Command{
	Use:   "trim <path>",
	Short: "Trim whitespace from secret values",
	Args:  cobra.ExactArgs(1),
	RunE:  runTrim,
}

func init() {
	trimCmd.Flags().String("config", ".vaultshift.yaml", "config file")
	trimCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	trimCmd.Flags().String("keys", "", "comma-separated keys to trim (default: all)")
	rootCmd.AddCommand(trimCmd)
}

func runTrim(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	keysRaw, _ := cmd.Flags().GetString("keys")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(vault.Config{
		Address: cfg.Source.Address,
		Token:   cfg.Source.Token,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	var keys []string
	if keysRaw != "" {
		for _, k := range strings.Split(keysRaw, ",") {
			if k = strings.TrimSpace(k); k != "" {
				keys = append(keys, k)
			}
		}
	}

	trimmer, err := vault.NewTrimmer(client, logger, dryRun, keys)
	if err != nil {
		return fmt.Errorf("trimmer: %w", err)
	}

	modified, err := trimmer.Trim(args[0])
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Println("[dry-run] would trim:", args[0])
	} else if modified {
		fmt.Println("trimmed:", args[0])
	} else {
		fmt.Println("no changes:", args[0])
	}
	return nil
}
