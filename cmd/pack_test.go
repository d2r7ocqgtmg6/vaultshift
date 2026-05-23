package cmd

import (
	"testing"
)

func TestPackCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "pack")
	if cmd == nil {
		t.Fatal("pack command not registered on root")
	}
}

func TestPackCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "pack")
	if cmd == nil {
		t.Fatal("pack command not found")
	}
	for _, flag := range []string{"config", "out", "dry-run"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestPackCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "pack")
	if cmd == nil {
		t.Fatal("pack command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"secret/data/foo"})
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestPackCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "pack")
	if cmd == nil {
		t.Fatal("pack command not found")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error when no paths provided")
	}
}
