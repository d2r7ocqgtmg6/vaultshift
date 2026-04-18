package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var annotateCmd = &cobra.Command{
	Use:   "annotate <path>",
	Short: "Add metadata annotations to a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runAnnotate,
}

func init() {
	AnnotateCmd := annotateCmd
	AnnotateCmd.Flags().StringSliceP("annotation", "a", nil, "Annotations in key=value format (repeatable)")
	AnnotateCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	AnnotateCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "Config file path")
	rootCmd.AddCommand(AnnotateCmd)
}

func runAnnotate(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	rawAnnotations, _ := cmd.Flags().GetStringSlice("annotation")
	if len(rawAnnotations) == 0 {
		return fmt.Errorf("at least one --annotation key=value is required")
	}

	annotations := make(map[string]string, len(rawAnnotations))
	for _, raw := range rawAnnotations {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid annotation %q: must be key=value", raw)
		}
		annotations[parts[0]] = parts[1]
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	logger, err := audit.New(cfg.AuditLog)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	annotator, err := vault.NewAnnotator(client, logger, annotations, dryRun)
	if err != nil {
		return err
	}

	path := args[0]
	if err := annotator.Annotate(path); err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would annotate %s\n", path)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "annotated %s\n", path)
	}
	return nil
}
