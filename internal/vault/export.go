package vault

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/drew/vaultshift/internal/audit"
)

// Exporter writes secrets from a Vault prefix to a local JSON file.
type Exporter struct {
	client *Client
	logger *audit.Logger
}

// NewExporter creates an Exporter. Returns error if client is nil.
func NewExporter(client *Client, logger *audit.Logger) (*Exporter, error) {
	if client == nil {
		return nil, fmt.Errorf("export: client is required")
	}
	return &Exporter{client: client, logger: logger}, nil
}

// Export lists all secrets under prefix and writes them to destPath as JSON.
// If dryRun is true, no file is written.
func (e *Exporter) Export(prefix, destPath string, dryRun bool) (int, error) {
	paths, err := listAll(e.client, prefix)
	if err != nil {
		return 0, fmt.Errorf("export: list error: %w", err)
	}

	data := make(map[string]map[string]interface{})
	for _, p := range paths {
		secret, err := e.client.ReadSecret(p)
		if err != nil {
			return 0, fmt.Errorf("export: read %s: %w", p, err)
		}
		data[p] = secret
		if e.logger != nil {
			e.logger.Log("export", p, nil)
		}
	}

	if dryRun {
		return len(data), nil
	}

	f, err := os.Create(destPath)
	if err != nil {
		return 0, fmt.Errorf("export: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return 0, fmt.Errorf("export: encode: %w", err)
	}

	return len(data), nil
}
