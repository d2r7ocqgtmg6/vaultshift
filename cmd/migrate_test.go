package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestMigrateCmd_RegisteredOnRoot(t *testing.T) {
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "migrate" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'migrate' subcommand to be registered on root")
	}
}

func TestMigrateCmd_HasExpectedFlags(t *testing.T) {
	_ = migrateCmd // ensure init() has run

	if rootCmd.PersistentFlags().Lookup("dry-run") == nil {
		t.Error("expected --dry-run flag on root command")
	}
	if rootCmd.PersistentFlags().Lookup("config") == nil {
		t.Error("expected --config flag on root command")
	}
	if rootCmd.PersistentFlags().Lookup("audit-log") == nil {
		t.Error("expected --audit-log flag on root command")
	}
}

func TestMigrateCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.AddCommand(migrateCmd)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Point to a non-existent config to trigger a load error.
	cfgFile = "/tmp/nonexistent-vaultshift-config-xyz.yaml"
	t.Cleanup(func() { cfgFile = "" })

	err := runMigrate(migrateCmd, []string{})
	if err == nil {
		t.Fatal("expected error when config file is missing, got nil")
	}
}
