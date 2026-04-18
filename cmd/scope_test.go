package cmd

import (
	"testing"
)

func TestScopeCmd_RegisteredOnRoot(t *testing.T) {
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Use == "scope <prefix> <path...>" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("scope command not registered on root")
	}
}

func TestScopeCmd_HasExpectedFlags(t *testing.T) {
	var scopeCmd = rootCmd.Commands()
	var cmd = findCmd(scopeCmd, "scope <prefix> <path...>")
	if cmd == nil {
		t.Fatal("scope command not found")
	}
	if cmd.Flags().Lookup("config") == nil {
		t.Error("missing --config flag")
	}
	if cmd.Flags().Lookup("list") == nil {
		t.Error("missing --list flag")
	}
}

func TestScopeCmd_MissingConfig_ReturnsError(t *testing.T) {
	var scopeCmd = rootCmd.Commands()
	var cmd = findCmd(scopeCmd, "scope <prefix> <path...>")
	if cmd == nil {
		t.Fatal("scope command not found")
	}
	_ = cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"secret/myapp", "secret/myapp/db"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}
