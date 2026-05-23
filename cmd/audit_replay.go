package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var auditReplayCmd = &cobra.Command{
	Use:   "audit-replay <audit-file>",
	Short: "Re-apply write operations from an audit log",
	Args:  cobra.ExactArgs(1),
	RunE:  runAuditReplay,
}

func init() {
	auditReplayCmd.Flags().String("config", ".vaultshift.yaml", "config file path")
	auditReplayCmd.Flags().Bool("dry-run", false, "preview replay without writing")
	auditReplayCmd.Flags().String("audit-log", "", "path to write audit output")
	rootCmd.AddCommand(auditReplayCmd)
}

func runAuditReplay(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditLog, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(auditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.DestAddress, cfg.DestToken)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	replayer, err := vault.NewReplayer(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("replayer: %w", err)
	}

	n, errs := replayer.Replay(args[0])
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "error: %v\n", e)
	}
	fmt.Printf("replayed %d secret(s)\n", n)
	if len(errs) > 0 {
		return fmt.Errorf("%d replay error(s)", len(errs))
	}
	return nil
}
