package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultshift/internal/config"
	"vaultshift/internal/vault"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify migrated secrets match the source namespace",
	RunE:  runVerify,
}

func init() {
	verifyCmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
	verifyCmd.Flags().StringSlice("paths", nil, "Explicit secret paths to verify (overrides prefix scan)")
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	paths, _ := cmd.Flags().GetStringSlice("paths")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	src, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}

	dst, err := vault.New(cfg.Destination.Address, cfg.Destination.Token)
	if err != nil {
		return fmt.Errorf("destination client: %w", err)
	}

	if len(paths) == 0 {
		lister := vault.NewLister(src)
		paths, err = lister.List(cfg.Source.Prefix)
		if err != nil {
			return fmt.Errorf("list secrets: %w", err)
		}
	}

	verifier := vault.NewVerifier(src, dst)
	result, err := verifier.Verify(paths)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Matched:  %d\n", len(result.Matched))
	fmt.Fprintf(os.Stdout, "Missing:  %d\n", len(result.Missing))
	fmt.Fprintf(os.Stdout, "Mismatch: %d\n", len(result.Mismatch))

	for _, p := range result.Missing {
		fmt.Fprintf(os.Stdout, "  [MISSING]  %s\n", p)
	}
	for _, p := range result.Mismatch {
		fmt.Fprintf(os.Stdout, "  [MISMATCH] %s\n", p)
	}

	if len(result.Missing)+len(result.Mismatch) > 0 {
		return fmt.Errorf("verification failed: %d missing, %d mismatched",
			len(result.Missing), len(result.Mismatch))
	}
	return nil
}
