package cmd

import (
	"testing"
)

func TestStashCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "stash")
	if cmd == nil {
		t.Fatal("stash command not registered on root")
	}
}

func TestStashCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "stash")
	if cmd == nil {
		t.Fatal("stash command not found")
	}
	for _, flag := range []string{"config", "out", "dry-run"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestStashCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "stash")
	if cmd == nil {
		t.Fatal("stash command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/config.yaml")
	err := cmd.RunE(cmd, []string{"secret/a"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestStashCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "stash")
	if cmd == nil {
		t.Fatal("stash command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no paths provided")
	}
}
