package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var maskCmd = &cobra.Command{
	Use:   "mask <path>",
	Short: "Replace specified secret keys with a mask value",
	Args:  cobra.ExactArgs(1),
	RunE:  runMask,
}

func init() {
	maskCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file")
	maskCmd.Flags().StringSlice("keys", nil, "keys to mask (comma-separated)")
	maskCmd.Flags().String("mask", "***", "replacement value")
	maskCmd.Flags().Bool("dry-run", false, "preview without writing")
	rootCmd.AddCommand(maskCmd)
}

func runMask(cmd *cobra.Command, args []string) error {
	cfgFile, _ := cmd.Flags().GetString("config")
	keys, _ := cmd.Flags().GetStringSlice("keys")
	maskVal, _ := cmd.Flags().GetString("mask")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return err
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return err
	}

	// support comma-separated keys passed as a single flag value
	var flatKeys []string
	for _, k := range keys {
		for _, part := range strings.Split(k, ",") {
			if t := strings.TrimSpace(part); t != "" {
				flatKeys = append(flatKeys, t)
			}
		}
	}

	masker, err := vault.NewMasker(client, logger, flatKeys, maskVal, dryRun)
	if err != nil {
		return err
	}

	return masker.Mask(args[0])
}
