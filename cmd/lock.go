package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/vault"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Manage advisory migration locks in Vault",
}

var lockAcquireCmd = &cobra.Command{
	Use:   "acquire",
	Short: "Acquire a migration lock at the configured lock path",
	RunE:  runLockAcquire,
}

var lockReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release the migration lock at the configured lock path",
	RunE:  runLockRelease,
}

var lockStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check whether the migration lock is currently held",
	RunE:  runLockStatus,
}

func init() {
	lockCmd.PersistentFlags().String("config", ".vaultshift.yaml", "path to config file")
	lockCmd.PersistentFlags().String("lock-path", "secret/data/vaultshift/lock", "Vault path used for the advisory lock")
	lockCmd.PersistentFlags().String("identity", "", "identity label written into the lock (defaults to hostname)")

	lockCmd.AddCommand(lockAcquireCmd, lockReleaseCmd, lockStatusCmd)
	rootCmd.AddCommand(lockCmd)
}

func resolveLocker(cmd *cobra.Command) (*vault.Locker, error) {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	lockPath, _ := cmd.Flags().GetString("lock-path")
	identity, _ := cmd.Flags().GetString("identity")
	if identity == "" {
		identity, _ = os.Hostname()
	}

	return vault.NewLocker(client, lockPath, identity)
}

func runLockAcquire(cmd *cobra.Command, _ []string) error {
	locker, err := resolveLocker(cmd)
	if err != nil {
		return err
	}
	if err := locker.Acquire(); err != nil {
		return fmt.Errorf("acquire: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "lock acquired")
	return nil
}

func runLockRelease(cmd *cobra.Command, _ []string) error {
	locker, err := resolveLocker(cmd)
	if err != nil {
		return err
	}
	if err := locker.Release(); err != nil {
		return fmt.Errorf("release: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "lock released")
	return nil
}

func runLockStatus(cmd *cobra.Command, _ []string) error {
	locker, err := resolveLocker(cmd)
	if err != nil {
		return err
	}
	locked, err := locker.IsLocked()
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if locked {
		fmt.Fprintln(cmd.OutOrStdout(), "status: locked")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "status: free")
	}
	return nil
}
