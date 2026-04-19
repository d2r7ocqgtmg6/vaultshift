package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var protectCmd = &cobra.Command{
	Use:   "protect <path>",
	Short: "Mark a secret as protected to prevent accidental modification",
	Args:  cobra.ExactArgs(1),
	RunE:  runProtect,
}

var unprotectCmd = &cobra.Command{
	Use:   "unprotect <path>",
	Short: "Remove the protection flag from a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnprotect,
}

func init() {
	for _, cmd := range []*cobra.Command{protectCmd, unprotectCmd} {
		cmd.Flags().String("config", ".vaultshift.yaml", "Path to config file")
		cmd.Flags().Bool("dry-run", false, "Preview without making changes")
		cmd.Flags().String("audit-log", "", "Path to audit log file")
	}
	rootCmd.AddCommand(protectCmd)
	rootCmd.AddCommand(unprotectCmd)
}

func buildProtector(cmd *cobra.Command, dryRun bool) (*vault.Protector, error) {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	c, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	logPath, _ := cmd.Flags().GetString("audit-log")
	l, err := audit.New(logPath)
	if err != nil {
		return nil, fmt.Errorf("audit logger: %w", err)
	}
	return vault.NewProtector(c, l, dryRun)
}

func runProtect(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	p, err := buildProtector(cmd, dryRun)
	if err != nil {
		return err
	}
	if err := p.Protect(args[0]); err != nil {
		return fmt.Errorf("protect: %w", err)
	}
	fmt.Printf("protected: %s\n", args[0])
	return nil
}

func runUnprotect(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	p, err := buildProtector(cmd, dryRun)
	if err != nil {
		return err
	}
	if err := p.Unprotect(args[0]); err != nil {
		return fmt.Errorf("unprotect: %w", err)
	}
	fmt.Printf("unprotected: %s\n", args[0])
	return nil
}
