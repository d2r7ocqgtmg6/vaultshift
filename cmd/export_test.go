package cmd

import (
	"testing"
)

func TestExportCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "export [prefix] [dest-file]" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("export command not registered on root")
	}
}

func TestExportCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "dry-run", "audit-log"}
	for _, f := range flags {
		if exportCmd.Flags().Lookup(f) == nil {
			t.Errorf("missing flag: %s", f)
		}
	}
}

func TestExportCmd_MissingConfig_ReturnsError(t *testing.T) {
	exportCmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := exportCmd.RunE(exportCmd, []string{"secret/prefix", "/tmp/out.json"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestExportCmd_RequiresTwoArgs(t *testing.T) {
	err := exportCmd.Args(exportCmd, []string{"only-one"})
	if err == nil {
		t.Fatal("expected error for wrong arg count")
	}
}
