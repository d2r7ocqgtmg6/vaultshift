package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var traceCmd = &cobra.Command{
	Use:   "trace [path...]",
	Short: "Trace read latency for one or more secret paths",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTrace,
}

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().StringP("config", "c", ".vaultshift.yaml", "config file path")
	traceCmd.Flags().StringP("log", "l", "", "audit log output file (default stdout)")
}

func runTrace(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	logPath, _ := cmd.Flags().GetString("log")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	logger, err := audit.New(logPath)
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	tracer, err := vault.NewTracer(client, logger)
	if err != nil {
		return fmt.Errorf("tracer: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tOP\tDURATION(ms)\tERROR")
	for _, path := range args {
		e := tracer.Trace(path)
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", e.Path, e.Operation, e.Duration, e.Error)
	}
	return w.Flush()
}
