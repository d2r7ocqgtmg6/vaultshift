package cmd

import (
	"testing"
)

func TestLintCmd_RegisteredOnRoot(t *testing.T) {
	if c := findCmd(rootCmd, "lint"); c == nil {
		t.Fatal("expected 'lint' command to be registered on root")
	}
}

func TestLintCmd_HasExpectedFlags(t *testing.T) {
	c := findCmd(rootCmd, "lint")
	if c == nil {
		t.Fatal("lint command not found")
	}
	for _, flag := range []string{"config", "no-empty", "no-uppercase"} {
		if c.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestLintCmd_MissingConfig_ReturnsError(t *testing.T) {
	c := findCmd(rootCmd, "lint")
	if c == nil {
		t.Fatal("lint command not found")
	}
	c.Flags().Set("config", "/nonexistent/path.yaml")
	err := c.RunE(c, []string{"secret/data/app"})
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
}

func TestLintCmd_RequiresArg(t *testing.T) {
	c := findCmd(rootCmd, "lint")
	if c == nil {
		t.Fatal("lint command not found")
	}
	err := c.Args(c, []string{})
	if err == nil {
		t.Fatal("expected error when no path argument is provided")
	}
}
