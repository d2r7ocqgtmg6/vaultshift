package vault

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Unpacker reads a packed JSON file and writes secrets back into Vault.
type Unpacker struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewUnpacker creates a new Unpacker. Returns an error if client or logger is nil.
func NewUnpacker(client *Client, logger *audit.Logger, dryRun bool) (*Unpacker, error) {
	if client == nil {
		return nil, fmt.Errorf("unpack: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("unpack: logger is required")
	}
	return &Unpacker{client: client, logger: logger, dryRun: dryRun}, nil
}

// Unpack reads secrets from the given file path and writes them to Vault.
// In dry-run mode no writes are performed.
func (u *Unpacker) Unpack(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unpack: read file: %w", err)
	}

	var entries map[string]map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("unpack: parse file: %w", err)
	}

	for path, secrets := range entries {
		if u.dryRun {
			u.logger.Log("unpack", map[string]interface{}{
				"path":    path,
				"dry_run": true,
				"keys":    len(secrets),
			})
			continue
		}
		if err := u.client.WriteSecret(path, secrets); err != nil {
			u.logger.Log("unpack_error", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
			return fmt.Errorf("unpack: write %s: %w", path, err)
		}
		u.logger.Log("unpack", map[string]interface{}{
			"path": path,
			"keys": len(secrets),
		})
	}
	return nil
}
