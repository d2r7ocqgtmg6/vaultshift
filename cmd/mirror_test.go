package cmd

import (
	"testing"
)

func TestMirrorCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "mirror")
	if cmd == nil {
		t.Fatal("mirror command not registered on root")
	}
}

func TestMirrorCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "mirror")
	if cmd == nil {
		t.Fatal("mirror command not found")
	}
	for _, flag := range []string{"config", "dry-run", "audit-log"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestMirrorCmd_RequiresTwoArgs(t *testing.T) {
	cmd := findCmd(rootCmd, "mirror")
	if cmd == nil {
		t.Fatal("mirror command not found")
	}
	if err := cmd.Args(cmd, []string{"only-one"}); err == nil {
		t.Error("expected error when fewer than two args provided")
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err != nil {
		t.Errorf("unexpected error with two args: %v", err)
	}
}

func TestMirrorCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "mirror")
	if cmd == nil {
		t.Fatal("mirror command not found")
	}
	_ = cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"src/", "dst/"})
	if err == nil {
		t.Error("expected error for missing config file")
	}
}
