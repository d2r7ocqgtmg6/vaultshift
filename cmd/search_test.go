package cmd

import (
	"testing"
)

func TestSearchCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "search <prefix> <query>" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("search command not registered on root")
	}
}

func TestSearchCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "values", "output"}
	for _, f := range flags {
		if searchCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined", f)
		}
	}
}

func TestSearchCmd_MissingConfig_ReturnsError(t *testing.T) {
	searchCmd.Flags().Set("config", "/nonexistent/path.yaml")
	err := searchCmd.RunE(searchCmd, []string{"secret/", "query"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestSearchCmd_RequiresTwoArgs(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{"only-one"})
	if err == nil {
		t.Fatal("expected error when only one arg provided")
	}
}
