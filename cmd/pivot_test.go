package cmd

import (
	"testing"
)

func TestPivotCmd_RegisteredOnRoot(t *testing.T) {
	if findCmd(rootCmd, "pivot") == nil {
		t.Fatal("expected 'pivot' command to be registered on root")
	}
}

func TestPivotCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "pivot")
	if cmd == nil {
		t.Fatal("pivot command not found")
	}

	for _, flag := range []string{"config", "dry-run", "audit-log"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestPivotCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "pivot")
	if cmd == nil {
		t.Fatal("pivot command not found")
	}
	_ = cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"secret/foo"})
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
}

func TestPivotCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "pivot")
	if cmd == nil {
		t.Fatal("pivot command not found")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Fatal("expected error when no paths are supplied")
	}
}
