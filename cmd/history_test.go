package cmd

import (
	"testing"
)

func TestHistoryCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "history <secret-path>" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'history' command to be registered on root")
	}
}

func TestHistoryCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "mount", "data"}
	for _, f := range flags {
		if historyCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined", f)
		}
	}
}

func TestHistoryCmd_MissingConfig_ReturnsError(t *testing.T) {
	historyCmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := historyCmd.RunE(historyCmd, []string{"app/db"})
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
}

func TestHistoryCmd_RequiresArg(t *testing.T) {
	err := historyCmd.Args(historyCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no secret path argument is provided")
	}
}
