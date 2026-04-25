package cmd

import (
	"testing"
)

func TestReplicateCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "replicate")
	if cmd == nil {
		t.Fatal("replicate command not registered on root")
	}
}

func TestReplicateCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "replicate")
	if cmd == nil {
		t.Fatal("replicate command not found")
	}
	for _, flag := range []string{"config", "dry-run", "overwrite", "audit-log"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestReplicateCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "replicate")
	if cmd == nil {
		t.Fatal("replicate command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := cmd.RunE(cmd, []string{"secret/src/", "secret/dst/"})
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
}

func TestReplicateCmd_RequiresTwoArgs(t *testing.T) {
	cmd := findCmd(rootCmd, "replicate")
	if cmd == nil {
		t.Fatal("replicate command not found")
	}
	err := cmd.Args(cmd, []string{"only-one"})
	if err == nil {
		t.Fatal("expected error when fewer than two args provided")
	}
}
