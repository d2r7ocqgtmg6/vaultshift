package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var scoreCmd = &cobra.Command{
	Use:   "score <path>",
	Short: "Score the health of a secret at the given path",
	Args:  cobra.ExactArgs(1),
	RunE:  runScore,
}

func init() {
	scoreCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "path to config file")
	scoreCmd.Flags().StringSlice("required-keys", nil, "comma-separated list of required secret keys")
	scoreCmd.Flags().Int("max-age-days", 0, "flag age threshold in days (0 = disabled)")
	scoreCmd.Flags().String("audit-log", "", "path to audit log file (default: stdout)")
	rootCmd.AddCommand(scoreCmd)
}

func runScore(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	auditPath, _ := cmd.Flags().GetString("audit-log")
	requiredKeys, _ := cmd.Flags().GetStringSlice("required-keys")
	maxAge, _ := cmd.Flags().GetInt("max-age-days")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("score: load config: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("score: audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("score: vault client: %w", err)
	}

	scorer, err := vault.NewScorer(client, logger, maxAge, requiredKeys)
	if err != nil {
		return fmt.Errorf("score: %w", err)
	}

	result, err := scorer.Score(args[0])
	if err != nil {
		return fmt.Errorf("score: %w", err)
	}

	cmd.Printf("path:  %s\n", result.Path)
	cmd.Printf("score: %d/100\n", result.Score)
	if len(result.Issues) > 0 {
		cmd.Printf("issues:\n  - %s\n", strings.Join(result.Issues, "\n  - "))
	} else {
		cmd.Println("issues: none")
	}
	return nil
}
