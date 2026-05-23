package cmd

import (
	"testing"
)

func TestSwapCmd_RegisteredOnRoot(t *testing.T) {
	if c := findCmd(rootCmd, "swap"); c == nil {
		t.Fatal("swap command not registered on root")
	}
}

func TestSwapCmd_HasExpectedFlags(t *testing.T) {
	c := findCmd(rootCmd, "swap")
	if c == nil {
		t.Fatal("swap command not found")
	}
	for _, flag := range []string{"config", "dry-run", "audit-log"} {
		if c.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestSwapCmd_RequiresTwoArgs(t *testing.T) {
	c := findCmd(rootCmd, "swap")
	if c == nil {
		t.Fatal("swap command not found")
	}
	err := c.Args(c, []string{"only-one"})
	if err == nil {
		t.Fatal("expected error when fewer than two args provided")
	}
}

func TestSwapCmd_MissingConfig_ReturnsError(t *testing.T) {
	c := findCmd(rootCmd, "swap")
	if c == nil {
		t.Fatal("swap command not found")
	}
	_ = c.Flags().Set("config", "/nonexistent/path.yaml")
	err := c.RunE(c, []string{"secret/a", "secret/b"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}
