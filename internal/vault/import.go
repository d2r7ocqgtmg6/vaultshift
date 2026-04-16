package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/vaultshift/internal/audit"
)

// ImportEntry represents a single secret entry in an import file.
type ImportEntry struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

// Importer writes secrets from a JSON file into Vault.
type Importer struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewImporter creates a new Importer.
func NewImporter(client *Client, logger *audit.Logger, dryRun bool) (*Importer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Importer{client: client, logger: logger, dryRun: dryRun}, nil
}

// Import reads entries from the given file path and writes them to Vault.
func (i *Importer) Import(ctx context.Context, filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	var entries []ImportEntry
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return 0, fmt.Errorf("decode import file: %w", err)
	}

	count := 0
	for _, e := range entries {
		fields := map[string]interface{}{
			"path":    e.Path,
			"dry_run": i.dryRun,
		}
		if i.dryRun {
			i.logger.Log("import_skip", fields)
			count++
			continue
		}
		if err := i.client.WriteSecret(ctx, e.Path, e.Data); err != nil {
			fields["error"] = err.Error()
			i.logger.Log("import_error", fields)
			continue
		}
		i.logger.Log("import_ok", fields)
		count++
	}
	return count, nil
}
