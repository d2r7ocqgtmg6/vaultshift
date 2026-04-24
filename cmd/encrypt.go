package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultshift/internal/audit"
	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt <path>",
	Short: "Encrypt secret values at a Vault path using AES-GCM",
	Args:  cobra.ExactArgs(1),
	RunE:  runEncrypt,
}

func init() {
	encryptCmd.Flags().String("config", ".vaultshift.yaml", "path to config file")
	encryptCmd.Flags().String("key-hex", "", "32-byte AES key in hex (env: VAULTSHIFT_ENCRYPT_KEY)")
	encryptCmd.Flags().Bool("dry-run", false, "preview without writing")
	encryptCmd.Flags().String("audit-log", "", "path to audit log file (default: stdout)")
	rootCmd.AddCommand(encryptCmd)
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	keyHex, _ := cmd.Flags().GetString("key-hex")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	auditPath, _ := cmd.Flags().GetString("audit-log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("encrypt: load config: %w", err)
	}

	if keyHex == "" {
		return fmt.Errorf("encrypt: --key-hex is required")
	}

	key, err := decodeHexKey(keyHex)
	if err != nil {
		return fmt.Errorf("encrypt: invalid key: %w", err)
	}

	logger, err := audit.New(auditPath)
	if err != nil {
		return fmt.Errorf("encrypt: audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("encrypt: vault client: %w", err)
	}

	enc, err := vault.NewEncrypter(client, logger, key, dryRun)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := enc.Encrypt(args[0]); err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would encrypt values at %s\n", args[0])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "encrypted values at %s\n", args[0])
	}
	return nil
}

func decodeHexKey(h string) ([]byte, error) {
	if len(h)%2 != 0 {
		return nil, fmt.Errorf("odd-length hex string")
	}
	buf := make([]byte, len(h)/2)
	for i := 0; i < len(h); i += 2 {
		var b byte
		_, err := fmt.Sscanf(h[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		buf[i/2] = b
	}
	return buf, nil
}
