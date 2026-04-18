package cmd

import (
	"testing"
)

func TestValidateCmd_RegisteredOnRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "validate <path>" {
			return
		}
	}
	t.Fatal("validate command not registered on root")
}

func TestValidateCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"config", "dry-run", "required-keys", "no-empty-values"}
	for _, f := range flags {
		if validateCmd.Flags().Lookup(f) == nil {
			t.Errorf("missing flag: %s", f)
		}
	}
}

func TestValidateCmd_MissingConfig_ReturnsError(t *testing.T) {
	validateCmd.Flags().Set("config", "/nonexistent/path.yaml")
	validateCmd.Flags().Set("required-keys", "key1")
	err := validateCmd.RunE(validateCmd, []string{"secret/app"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestValidateCmd_RequiresArg(t *testing.T) {
	err := validateCmd.Args(validateCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no path argument provided")
	}
}
