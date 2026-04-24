package cmd

import (
	"fmt"
	"strings"

	"github.com/dreamsofcode-io/vaultshift/internal/audit"
	"github.com/dreamsofcode-io/vaultshift/internal/config"
	"github.com/dreamsofcode-io/vaultshift/internal/vault"
	"github.com/spf13/cobra"
)

var pivotCmd = &cobra.Command{
	Use:   "pivot [path...]",
	Short: "Pivot secrets from source to destination mount",
	Long:  "Reads each supplied path from the source Vault and writes it to the destination Vault under the same relative path.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPivot,
}

func init() {
	pivotCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	pivotCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	pivotCmd.Flags().String("audit-log", "", "path to audit log file (defaults to stdout)")
	rootCmd.AddCommand(pivotCmd)
}

func runPivot(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("pivot: load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("pivot: audit logger: %w", err)
	}

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("pivot: source client: %w", err)
	}
	dst, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("pivot: destination client: %w", err)
	}

	pivoter, err := vault.NewPivoter(src, dst, logger, dryRun)
	if err != nil {
		return fmt.Errorf("pivot: %w", err)
	}

	if err := pivoter.Pivot(args); err != nil {
		return fmt.Errorf("pivot: %w", err)
	}

	if dryRun {
		fmt.Printf("[dry-run] would pivot %d path(s): %s\n", len(args), strings.Join(args, ", "))
	} else {
		fmt.Printf("pivoted %d path(s)\n", len(args))
	}
	return nil
}
