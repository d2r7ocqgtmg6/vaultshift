package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Rekeyer renames all keys within a secret by applying a mapping.
type Rekeyer struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewRekeyer creates a new Rekeyer.
func NewRekeyer(client *Client, logger *audit.Logger, dryRun bool) (*Rekeyer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Rekeyer{client: client, logger: logger, dryRun: dryRun}, nil
}

// RekeyResult holds the outcome of a rekey operation.
type RekeyResult struct {
	Path    string
	Renamed map[string]string
	Skipped []string
}

// Rekey reads the secret at path and renames its data keys according to the
// provided mapping (oldKey -> newKey). Keys not in the mapping are left as-is.
func (r *Rekeyer) Rekey(path string, mapping map[string]string) (*RekeyResult, error) {
	secret, err := r.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	result := &RekeyResult{
		Path:    path,
		Renamed: make(map[string]string),
	}

	newData := make(map[string]interface{}, len(secret))
	for k, v := range secret {
		if newKey, ok := mapping[k]; ok {
			newData[newKey] = v
			result.Renamed[k] = newKey
		} else {
			newData[k] = v
			result.Skipped = append(result.Skipped, k)
		}
	}

	r.logger.Log("rekey", map[string]interface{}{
		"path":    path,
		"renamed": result.Renamed,
		"dry_run": r.dryRun,
	})

	if r.dryRun {
		return result, nil
	}

	if err := r.client.WriteSecret(path, newData); err != nil {
		return nil, fmt.Errorf("write %s: %w", path, err)
	}

	return result, nil
}
