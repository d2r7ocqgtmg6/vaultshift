package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findCmd(rootCmd, "score")
	require.NotNil(t, cmd, "expected 'score' command to be registered on root")
}

func TestScoreCmd_HasExpectedFlags(t *testing.T) {
	cmd := findCmd(rootCmd, "score")
	require.NotNil(t, cmd)

	assert.NotNil(t, cmd.Flags().Lookup("config"), "expected --config flag")
	assert.NotNil(t, cmd.Flags().Lookup("required-keys"), "expected --required-keys flag")
	assert.NotNil(t, cmd.Flags().Lookup("max-age-days"), "expected --max-age-days flag")
	assert.NotNil(t, cmd.Flags().Lookup("audit-log"), "expected --audit-log flag")
}

func TestScoreCmd_MissingConfig_ReturnsError(t *testing.T) {
	cmd := findCmd(rootCmd, "score")
	require.NotNil(t, cmd)

	cmd.Flags().Set("config", "/nonexistent/path.yaml") //nolint:errcheck
	err := cmd.RunE(cmd, []string{"secret/data/app"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "score")
}

func TestScoreCmd_RequiresArg(t *testing.T) {
	cmd := findCmd(rootCmd, "score")
	require.NotNil(t, cmd)

	err := cmd.Args(cmd, []string{})
	require.Error(t, err)
}
