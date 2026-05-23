package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wearevault/vaultshift/internal/audit"
	"github.com/wearevault/vaultshift/internal/config"
	"github.com/wearevault/vaultshift/internal/vault"
)

var digestCmd = &cobra.Command{
	Use:   "digest <prefix>",
	Short: "Compute SHA-256 digests for secrets under a prefix",
	Args:  cobra.ExactArgs(1),
	RunE:  runDigest,
}

func init() {
	digestCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	digestCmd.Flags().Bool("dry-run", false, "print digests without writing")
	rootCmd.AddCommand(digestCmd)
}

func runDigest(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
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

	digester, err := vault.NewDigester(client, logger)
	if err != nil {
		return fmt.Errorf("digester: %w", err)
	}

	prefix := args[0]
	results, err := digester.DigestAll(prefix)
	if err != nil {
		return fmt.Errorf("digest: %w", err)
	}

	for _, r := range results {
		fmt.Fprintf(os.Stdout, "%s\t%s\n", r.Digest, r.Path)
	}
	return nil
}
