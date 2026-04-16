package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultshift/internal/config"
	"github.com/vaultshift/internal/vault"
)

var searchCmd = &cobra.Command{
	Use:   "search <prefix> <query>",
	Short: "Search secrets by key or value under a prefix",
	Args:  cobra.ExactArgs(2),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().String("config", ".vaultshift.yaml", "config file path")
	searchCmd.Flags().Bool("values", false, "also search secret values")
	searchCmd.Flags().String("output", "text", "output format: text or json")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	searchValues, _ := cmd.Flags().GetBool("values")
	output, _ := cmd.Flags().GetString("output")
	prefix := args[0]
	query := args[1]

	client, err := vault.New(cfg.Source.Address, cfg.Source.Token)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	searcher, err := vault.NewSearcher(client)
	if err != nil {
		return err
	}

	results, err := searcher.Search(prefix, query, searchValues)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if output == "json" {
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	if len(results) == 0 {
		fmt.Println("no matches found")
		return nil
	}
	for _, r := range results {
		fmt.Printf("%s  [%s]\n", r.Path, r.Key)
	}
	return nil
}
