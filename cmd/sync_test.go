package cmd

import (
	"testing"
)

func TestSyncCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "sync <src-prefix> <dst-prefix>" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("sync command not registered on root")
	}
}

func TestSyncCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "dry-run"}
	for _, f := range flags {
		if syncCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined", f)
		}
	}
}

func TestSyncCmd_MissingConfig_ReturnsError(t *testing.T) {
	syncCmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint
	err := runSync(syncCmd, []string{"secret/src/", "secret/dst/"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestSyncCmd_RequiresTwoArgs(t *testing.T) {
	if syncCmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}
	if err := syncCmd.Args(syncCmd, []string{"only-one"}); err == nil {
		t.Fatal("expected error for single argument")
	}
	if err := syncCmd.Args(syncCmd, []string{"a", "b"}); err != nil {
		t.Fatalf("expected no error for two args: %v", err)
	}
}
