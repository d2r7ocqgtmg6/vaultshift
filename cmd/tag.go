package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultshift/vaultshift/internal/config"
	"github.com/vaultshift/vaultshift/internal/vault"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Annotate secret paths with key=value metadata tags",
	RunE:  runTag,
}

func init() {
	tagCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "Path to config file")
	tagCmd.Flags().StringP("prefix", "p", "", "Secret path prefix to tag")
	tagCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags in key=value format (repeatable)")
	tagCmd.Flags().Bool("dry-run", false, "Print annotated paths without writing")
	rootCmd.AddCommand(tagCmd)
}

func runTag(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	prefix, _ := cmd.Flags().GetString("prefix")
	rawTags, _ := cmd.Flags().GetStringSlice("tags")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	tagger, err := vault.NewTagger(rawTags)
	if err != nil {
		return fmt.Errorf("building tagger: %w", err)
	}

	srcClient, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("source vault client: %w", err)
	}

	paths, err := vault.ListSecrets(srcClient, prefix)
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	annotated := tagger.TagPaths(paths)

	if dryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] Annotated paths:")
		for path, tags := range annotated {
			fmt.Fprintf(os.Stdout, "  %s -> %v\n", path, tags)
		}
		return nil
	}

	for path, tags := range annotated {
		data, readErr := srcClient.ReadSecret(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "warn: could not read %s: %v\n", path, readErr)
			continue
		}
		merged := tagger.Annotate(data)
		if writeErr := srcClient.WriteSecret(path, merged); writeErr != nil {
			fmt.Fprintf(os.Stderr, "warn: could not write %s: %v\n", path, writeErr)
			continue
		}
		_ = tags
		fmt.Fprintf(os.Stdout, "tagged: %s\n", path)
	}
	return nil
}
