package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/vault"
)

var transformCmd = &cobra.Command{
	Use:   "transform <path>",
	Short: "Preview path and key transformations without migrating",
	Args:  cobra.ExactArgs(1),
	RunE:  runTransform,
}

func init() {
	transformCmd.Flags().String("strip-prefix", "", "Prefix to strip from secret path")
	transformCmd.Flags().String("add-prefix", "", "Prefix to prepend to secret path")
	transformCmd.Flags().StringToString("rename-keys", nil, "Key renames as old=new pairs (comma-separated)")
	rootCmd.AddCommand(transformCmd)
}

func runTransform(cmd *cobra.Command, args []string) error {
	path := args[0]

	stripPrefix, _ := cmd.Flags().GetString("strip-prefix")
	addPrefix, _ := cmd.Flags().GetString("add-prefix")
	renameKeys, _ := cmd.Flags().GetStringToString("rename-keys")

	rule := vault.TransformRule{
		StripPrefix: stripPrefix,
		AddPrefix:   addPrefix,
		RenameKeys:  renameKeys,
	}

	tr := vault.NewTransformer(rule)

	transformedPath, err := tr.TransformPath(path)
	if err != nil {
		return fmt.Errorf("transform path: %w", err)
	}

	result := map[string]interface{}{
		"original_path":    path,
		"transformed_path": transformedPath,
		"rename_keys":      renameKeys,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}
