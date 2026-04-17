package cmd

import (
	"testing"
)

func TestPinCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "pin <path>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("pin command not registered on root")
	}
}

func TestPinCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "dry-run", "audit-log"}
	for _, f := range flags {
		if pinCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined", f)
		}
	}
}

func TestPinCmd_MissingConfig_ReturnsError(t *testing.T) {
	rootCmd.SetArgs([]string{"pin", "--config", "nonexistent.yaml", "secret/data/test"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestPinCmd_RequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"pin"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when no path argument provided")
	}
}
