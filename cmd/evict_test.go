package cmd

import (
	"testing"
)

func TestEvictCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "evict")
	if cmd == nil {
		t.Fatal("evict command not registered on root")
	}
}

func TestEvictCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "evict")
	if cmd == nil {
		t.Fatal("evict command not found")
	}
	for _, flag := range []string{"dry-run", "config", "audit-log"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestEvictCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "evict")
	if cmd == nil {
		t.Fatal("evict command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"secret/data/app", "password"})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}

func TestEvictCmd_RequiresTwoArgs(t *testing.T) {
	cmd := findCmd(rootCmd, "evict")
	if cmd == nil {
		t.Fatal("evict command not found")
	}
	err := cmd.Args(cmd, []string{"only-one"})
	if err == nil {
		t.Fatal("expected error when fewer than 2 args provided")
	}
}
