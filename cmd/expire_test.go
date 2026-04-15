package cmd

import (
	"testing"
)

func TestExpireCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "expire" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expire command not registered on root")
	}
}

func TestExpireCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "meta-key", "prefix", "dry-run", "audit-log"}
	for _, name := range flags {
		if expireCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be defined", name)
		}
	}
}

func TestExpireCmd_MissingConfig_ReturnsError(t *testing.T) {
	expireCmd.Flags().Set("config", "/nonexistent/vaultshift.yaml") //nolint:errcheck
	err := expireCmd.RunE(expireCmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
