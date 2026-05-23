package cmd

import (
	"testing"
)

func TestDigestCmd_RegisteredOnRoot(t *testing.T) {
	if c := findCmd(rootCmd, "digest"); c == nil {
		t.Fatal("digest command not registered on root")
	}
}

func TestDigestCmd_HasExpectedFlags(t *testing.T) {
	c := findCmd(rootCmd, "digest")
	if c == nil {
		t.Fatal("digest command not found")
	}
	if c.Flags().Lookup("config") == nil {
		t.Error("expected --config flag")
	}
	if c.Flags().Lookup("dry-run") == nil {
		t.Error("expected --dry-run flag")
	}
}

func TestDigestCmd_MissingConfig_ReturnsError(t *testing.T) {
	c := findCmd(rootCmd, "digest")
	if c == nil {
		t.Fatal("digest command not found")
	}
	_ = c.Flags().Set("config", "/nonexistent/path.yaml")
	err := c.RunE(c, []string{"secret/"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestDigestCmd_RequiresArg(t *testing.T) {
	c := findCmd(rootCmd, "digest")
	if c == nil {
		t.Fatal("digest command not found")
	}
	err := c.Args(c, []string{})
	if err == nil {
		t.Fatal("expected error when no arg provided")
	}
}
