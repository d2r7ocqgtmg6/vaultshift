package cmd

import (
	"testing"
)

func TestRotateCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "rotate <path>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rotate command not registered on root")
	}
}

func TestRotateCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "dry-run", "audit-log"}
	for _, f := range flags {
		if rotateCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s", f)
		}
	}
}

func TestRotateCmd_MissingConfig_ReturnsError(t *testing.T) {
	rotateCmd.Flags().Set("config", "/nonexistent/path.yaml")
	defer rotateCmd.Flags().Set("config", ".vaultshift.yaml")
	err := rotateCmd.RunE(rotateCmd, []string{"secret/data/test"})
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestRotateCmd_RequiresArg(t *testing.T) {
	err := rotateCmd.Args(rotateCmd, []string{})
	if err == nil {
		t.Error("expected error when no path arg provided")
	}
}
