package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var syncCmd = &cobra.Command{
	Use:   "sync <src-prefix> <dst-prefix>",
	Short: "Bidirectionally sync secrets between two prefixes",
	Args:  cobra.ExactArgs(2),
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	syncCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	srcPrefix := args[0]
	dstPrefix := args[1]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("src client: %w", err)
	}
	dst, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("dst client: %w", err)
	}

	syncer, err := vault.NewSyncer(src, dst, logger, dryRun)
	if err != nil {
		return fmt.Errorf("syncer: %w", err)
	}

	result, err := syncer.Sync(srcPrefix, dstPrefix)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	for _, p := range result.SrcToDst {
		fmt.Fprintf(os.Stdout, "[src->dst] %s\n", p)
	}
	for _, p := range result.DstToSrc {
		fmt.Fprintf(os.Stdout, "[dst->src] %s\n", p)
	}
	for _, e := range result.Errors {
		fmt.Fprintf(os.Stderr, "error: %s\n", e)
	}
	if dryRun {
		fmt.Println("dry-run: no changes written")
	}
	return nil
}
