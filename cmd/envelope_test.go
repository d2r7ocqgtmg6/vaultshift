package cmd

import (
	"testing"
)

func TestEnvelopeCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "envelope")
	if cmd == nil {
		t.Fatal("envelope command not registered on root")
	}
}

func TestEnvelopeCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "envelope")
	if cmd == nil {
		t.Fatal("envelope command not found")
	}

	for _, flag := range []string{"config", "out", "dry-run"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestEnvelopeCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "envelope")
	if cmd == nil {
		t.Fatal("envelope command not found")
	}

	cmd.Flags().Set("config", "/nonexistent/config.yaml")
	err := cmd.RunE(cmd, []string{"secret/data/foo"})
	if err == nil {
		t.Fatal("expected error with missing config")
	}
}

func TestEnvelopeCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "envelope")
	if cmd == nil {
		t.Fatal("envelope command not found")
	}

	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no paths provided")
	}
}
