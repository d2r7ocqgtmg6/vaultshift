package cmd

import "github.com/spf13/cobra"

// findCmd locates a subcommand by Use string within a slice.
func findCmd(cmds []*cobra.Command, use string) *cobra.Command {
	for _, c := range cmds {
		if c.Use == use {
			return c
		}
	}
	return nil
}
