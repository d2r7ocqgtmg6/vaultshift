package cmd

import (
	"testing"
)

func TestPromoteCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "promote" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("promote command not registered on root")
	}
}

func TestPromoteCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "src-prefix", "dst-prefix", "skip-existing", "dry-run"}
	for _, name := range flags {
		if promoteCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be defined", name)
		}
	}
}

func TestPromoteCmd_MissingConfig_ReturnsError(t *testing.T) {
	rootCmd.SetArgs([]string{
		"promote",
		"--config", "/nonexistent/path.yaml",
		"--src-prefix", "secret/staging/",
		"--dst-prefix", "secret/prod/",
	})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
