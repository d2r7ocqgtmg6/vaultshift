package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "filter" {
			found = true
			break
		}
	}
	assert.True(t, found, "filter command should be registered on root")
}

func TestFilterCmd_HasExpectedFlags(t *testing.T) {
	cmd := filterCmd
	require.NotNil(t, cmd.Flags().Lookup("config"), "--config flag expected")
	require.NotNil(t, cmd.Flags().Lookup("prefix"), "--prefix flag expected")
}

func TestFilterCmd_MissingConfig_ReturnsError(t *testing.T) {
	rootCmd.SetArgs([]string{"filter", "--config", "/nonexistent/path.yaml"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
