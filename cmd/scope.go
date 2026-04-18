package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

func init() {
	scopeCmd := &cobra.Command{
		Use:   "scope <prefix> <path...>",
		Short: "Validate and filter paths within a namespace scope",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runScope,
	}
	scopeCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file")
	scopeCmd.Flags().Bool("list", false, "list only paths within scope")
	rootCmd.AddCommand(scopeCmd)
}

func runScope(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	listOnly, _ := cmd.Flags().GetBool("list")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	prefix := args[0]
	paths := args[1:]

	scoper, err := vault.NewScoper(client, prefix)
	if err != nil {
		return fmt.Errorf("scoper: %w", err)
	}

	if listOnly {
		scoped := scoper.ScopedPaths(paths)
		if len(scoped) == 0 {
			fmt.Println("no paths within scope")
			return nil
		}
		fmt.Println(strings.Join(scoped, "\n"))
		return nil
	}

	for _, p := range paths {
		if _, err := scoper.Enforce(p); err != nil {
			fmt.Printf("OUT_OF_SCOPE  %s\n", p)
		} else {
			fmt.Printf("IN_SCOPE      %s\n", p)
		}
	}
	return nil
}
