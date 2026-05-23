package cmd

import (
	"testing"
)

func TestAuditReplayCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-replay")
	if cmd == nil {
		t.Fatal("audit-replay command not registered on root")
	}
}

func TestAuditReplayCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-replay")
	if cmd == nil {
		t.Fatal("audit-replay command not found")
	}
	for _, flag := range []string{"config", "dry-run", "audit-log"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestAuditReplayCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-replay")
	if cmd == nil {
		t.Fatal("audit-replay command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"/tmp/audit.jsonl"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestAuditReplayCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-replay")
	if cmd == nil {
		t.Fatal("audit-replay command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
