package cmd

import (
	"testing"
)

func TestGraphCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "graph")
	if cmd == nil {
		t.Fatal("graph command not registered on root")
	}
}

func TestGraphCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "graph")
	if cmd == nil {
		t.Fatal("graph command not found")
	}
	flags := []string{"config", "output"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be registered", f)
		}
	}
}

func TestGraphCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "graph")
	if cmd == nil {
		t.Fatal("graph command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := cmd.RunE(cmd, []string{"secret/metadata/test"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestGraphCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "graph")
	if cmd == nil {
		t.Fatal("graph command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
