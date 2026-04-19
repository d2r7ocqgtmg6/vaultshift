package cmd

import (
	"testing"
)

func TestCheckpointCmd_RegisteredOnRoot(t *testing.T) {
	if findCmd(rootCmd, "checkpoint") == nil {
		t.Fatal("checkpoint command not registered on root")
	}
}

func TestCheckpointCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "checkpoint")
	if cmd == nil {
		t.Fatal("checkpoint command not found")
	}
	for _, flag := range []string{"config", "dry-run"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("missing flag: %s", flag)
		}
	}
}

func TestCheckpointCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "checkpoint")
	if cmd == nil {
		t.Fatal("checkpoint command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/config.yaml")
	err := cmd.RunE(cmd, []string{"secret/app", "/tmp/cp.json"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestCheckpointCmd_RequiresTwoArgs(t *testing.T) {
	cmd := findCmd(rootCmd, "checkpoint")
	if cmd == nil {
		t.Fatal("checkpoint command not found")
	}
	if err := cmd.Args(cmd, []string{"only-one"}); err == nil {
		t.Fatal("expected error for single argument")
	}
}
