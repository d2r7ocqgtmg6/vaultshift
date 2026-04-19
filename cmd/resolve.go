package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <path> <key>",
	Short: "Resolve a secret reference to its value",
	Args:  cobra.ExactArgs(2),
	RunE:  runResolve,
}

func init() {
	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file")
	resolveCmd.Flags().Bool("dry-run", false, "print resolved value without writing anywhere")
	resolveCmd.Flags().Bool("map", false, "resolve all keys at path and print as map")
}

func runResolve(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	asMap, _ := cmd.Flags().GetBool("map")

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

	r, err := vault.NewResolver(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("resolver: %w", err)
	}

	path := args[0]

	if asMap {
		m, err := r.ResolveMap(path)
		if err != nil {
			return err
		}
		for k, v := range m {
			fmt.Fprintf(os.Stdout, "%s=%s\n", k, v)
		}
		return nil
	}

	key := args[1]
	val, err := r.Resolve(path, key)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, val)
	return nil
}
