package cmd

import (
	"testing"
)

func TestPolicyCmd_RegisteredOnRoot(t *testing.T) {
	if findCmd(rootCmd, "policy") == nil {
		t.Fatal("policy command not registered on root")
	}
}

func TestPolicyCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "policy")
	if cmd == nil {
		t.Fatal("policy command not found")
	}
	if cmd.Flags().Lookup("config") == nil {
		t.Error("missing --config flag")
	}
	if cmd.Flags().Lookup("audit-log") == nil {
		t.Error("missing --audit-log flag")
	}
}

func TestPolicyCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "policy")
	if cmd == nil {
		t.Fatal("policy command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"secret/data/foo"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestPolicyCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "policy")
	if cmd == nil {
		t.Fatal("policy command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no paths provided")
	}
}
