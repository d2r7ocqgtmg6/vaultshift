package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var batchCmd = &cobra.Command{
	Use:   "batch <secrets-json-file>",
	Short: "Write or delete multiple secrets in bulk from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatch,
}

func init() {
	batchCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	batchCmd.Flags().Bool("dry-run", false, "preview actions without applying them")
	batchCmd.Flags().Bool("delete", false, "delete the listed paths instead of writing")
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	deleteMode, _ := cmd.Flags().GetBool("delete")

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

	batcher, err := vault.NewBatcher(client, logger, dryRun)
	if err != nil {
		return fmt.Errorf("batcher: %w", err)
	}

	raw, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if deleteMode {
		var paths []string
		if err := json.Unmarshal(raw, &paths); err != nil {
			return fmt.Errorf("parse paths: %w", err)
		}
		results := batcher.DeleteAll(paths)
		for _, r := range results {
			if !r.Success {
				fmt.Fprintf(os.Stderr, "delete failed: %s: %v\n", r.Path, r.Error)
			}
		}
		return nil
	}

	var secrets map[string]map[string]interface{}
	if err := json.Unmarshal(raw, &secrets); err != nil {
		return fmt.Errorf("parse secrets: %w", err)
	}
	results := batcher.WriteAll(secrets)
	for _, r := range results {
		if !r.Success {
			fmt.Fprintf(os.Stderr, "write failed: %s: %v\n", r.Path, r.Error)
		}
	}
	return nil
}
