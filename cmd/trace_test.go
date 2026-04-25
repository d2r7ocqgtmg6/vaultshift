package cmd

import (
	"testing"
)

func TestTraceCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "trace")
	if cmd == nil {
		t.Fatal("expected trace command to be registered on root")
	}
}

func TestTraceCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "trace")
	if cmd == nil {
		t.Fatal("trace command not found")
	}

	flags := []string{"config", "log"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined", f)
		}
	}
}

func TestTraceCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "trace")
	if cmd == nil {
		t.Fatal("trace command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := cmd.RunE(cmd, []string{"secret/data/foo"})
	if err == nil {
		t.Error("expected error when config file is missing")
	}
}

func TestTraceCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "trace")
	if cmd == nil {
		t.Fatal("trace command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error when no path arguments provided")
	}
}
