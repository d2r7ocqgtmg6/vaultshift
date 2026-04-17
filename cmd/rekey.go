package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var rekeyCmd = &cobra.Command{
	Use:   "rekey <path>",
	Short: "Rename keys within a secret at the given path",
	Long: `Rekey reads a secret at the specified path and renames one or more
keys within its data according to the provided mapping (oldKey=newKey).
Use --dry-run to preview changes without writing to Vault.`,
	Args: cobra.ExactArgs(1),
	RunE: runRekey,
}

func init() {
	rekeyCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "Path to config file")
	rekeyCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	rekeyCmd.Flags().StringSliceP("map", "m", nil, "Key mappings in oldKey=newKey format (repeatable)")
	rekeyCmd.Flags().String("audit-log", "", "Path to audit log file (defaults to stdout)")
	_ = rekeyCmd.MarkFlagRequired("map")
	rootCmd.AddCommand(rekeyCmd)
}

func runRekey(cmd *cobra.Command, args []string) error {
	path := args[0]

	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	mappings, _ := cmd.Flags().GetStringSlice("map")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("creating audit logger: %w", err)
	}

	client, err := vault.New(vault.Config{
		Address: cfg.Source.Address,
		Token:   cfg.Source.Token,
		Mount:   cfg.Source.Mount,
	})
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	// Parse mappings from "oldKey=newKey" format.
	keyMap := make(map[string]string, len(mappings))
	for _, m := range mappings {
		var oldKey, newKey string
		n, _ := fmt.Sscanf(m, "%s", &oldKey)
		if n == 0 {
			return fmt.Errorf("invalid mapping %q: expected oldKey=newKey", m)
		}
		for i, ch := range m {
			if ch == '=' {
				oldKey = m[:i]
				newKey = m[i+1:]
				break
			}
		}
		if oldKey == "" || newKey == "" {
			return fmt.Errorf("invalid mapping %q: expected oldKey=newKey", m)
		}
		keyMap[oldKey] = newKey
	}

	rekeyer, err := vault.NewRekeyer(client, logger, keyMap, dryRun)
	if err != nil {
		return fmt.Errorf("creating rekeyer: %w", err)
	}

	result, err := rekeyer.Rekey(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("rekeying secret: %w", err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would rekey %d key(s) at %s\n", len(result.Renamed), path)
		for old, nw := range result.Renamed {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", old, nw)
		}
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "rekeyed %d key(s) at %s\n", len(result.Renamed), path)
	return nil
}
