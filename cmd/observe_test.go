package cmd

import (
	"testing"
)

func TestObserveCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "observe")
	if cmd == nil {
		t.Fatal("observe command not registered on root")
	}
}

func TestObserveCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "observe")
	if cmd == nil {
		t.Fatal("observe command not found")
	}
	for _, flag := range []string{"config", "dry-run"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestObserveCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "observe")
	if cmd == nil {
		t.Fatal("observe command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := cmd.RunE(cmd, []string{"secret/data/test"})
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
}

func TestObserveCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "observe")
	if cmd == nil {
		t.Fatal("observe command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
