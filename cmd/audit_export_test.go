package cmd

import (
	"testing"
)

func TestAuditExportCmd_RegisteredOnRoot(t *testing.T) {
	if findCmd(rootCmd, "audit-export") == nil {
		t.Fatal("audit-export command not registered on root")
	}
}

func TestAuditExportCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-export")
	if cmd == nil {
		t.Fatal("audit-export command not found")
	}
	for _, flag := range []string{"config", "dry-run", "operation"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to be defined", flag)
		}
	}
}

func TestAuditExportCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-export")
	if cmd == nil {
		t.Fatal("audit-export command not found")
	}
	cmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := cmd.RunE(cmd, []string{"/tmp/out.jsonl"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestAuditExportCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "audit-export")
	if cmd == nil {
		t.Fatal("audit-export command not found")
	}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no dest arg provided")
	}
}
