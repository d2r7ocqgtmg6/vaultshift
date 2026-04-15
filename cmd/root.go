package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	dryRun   bool
	auditLog string
)

var rootCmd = &cobra.Command{
	Use:   "vaultshift",
	Short: "Migrate secrets between HashiCorp Vault namespaces",
	Long: `vaultshift is a CLI tool to migrate secrets between HashiCorp Vault
namespaces with dry-run and audit logging support.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .vaultshift.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview migration without writing secrets")
	rootCmd.PersistentFlags().StringVar(&auditLog, "audit-log", "", "path to audit log file (default: stdout)")
}
