package cmd

import (
	"testing"
)

func TestRevertCmd_RegisteredOnRoot(t *testing.T) {
	if c := findCmd(rootCmd, "revert"); c == nil {
		t.Fatal("expected revert command to be registered on root")
	}
}

func TestRevertCmd_HasExpectedFlags(t *testing.T) {
	c := findCmd(rootCmd, "revert")
	if c == nil {
		t.Fatal("revert command not found")
	}
	for _, flag := range []string{"config", "dry-run", "audit-log"} {
		if c.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestRevertCmd_MissingConfig_ReturnsError(t *testing.T) {
	c := findCmd(rootCmd, "revert")
	if c == nil {
		t.Fatal("revert command not found")
	}
	c.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := c.RunE(c, []string{"/tmp/snap.json"})
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRevertCmd_RequiresArg(t *testing.T) {
	c := findCmd(rootCmd, "revert")
	if c == nil {
		t.Fatal("revert command not found")
	}
	err := c.Args(c, []string{})
	if err == nil {
		t.Fatal("expected error when no snapshot arg provided")
	}
}
