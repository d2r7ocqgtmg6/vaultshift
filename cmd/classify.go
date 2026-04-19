package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var classifyCmd = &cobra.Command{
	Use:   "classify [path]",
	Short: "Classify a secret by inspecting its key names against configured rules",
	Args:  cobra.ExactArgs(1),
	RunE:  runClassify,
}

func init() {
	classifyCmd.Flags().String("config", ".vaultshift.yaml", "config file")
	classifyCmd.Flags().StringToString("rules", map[string]string{"password": "sensitive", "token": "sensitive", "key": "restricted"}, "key substring to label mappings")
	rootCmd.AddCommand(classifyCmd)
}

func runClassify(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	rules, _ := cmd.Flags().GetStringToString("rules")
	path := args[0]

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

	classifier, err := vault.NewClassifier(client, logger, rules)
	if err != nil {
		return fmt.Errorf("classifier: %w", err)
	}

	result, err := classifier.Classify(path)
	if err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "path=%s label=%s\n", result.Path, result.Label)
	return nil
}
