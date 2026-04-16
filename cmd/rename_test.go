package cmd

import (
	"testing"
)

func TestRenameCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range RootCmd.Commands() {
		if c.Use == "rename <src> <dst>" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("rename command not registered on root")
	}
}

func TestRenameCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "dry-run"}
	for _, f := range flags {
		if renameCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined", f)
		}
	}
}

func TestRenameCmd_MissingConfig_ReturnsError(t *testing.T) {
	renameCmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := renameCmd.RunE(renameCmd, []string{"secret/data/old", "secret/data/new"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRenameCmd_RequiresTwoArgs(t *testing.T) {
	if err := renameCmd.Args(renameCmd, []string{"only-one"}); err == nil {
		t.Fatal("expected error when only one arg provided")
	}
}
