package vault

import (
	"encoding/json"
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Truncater truncates secret values to a maximum length.
type Truncater struct {
	client    *Client
	logger    *audit.Logger
	maxLen    int
	dryRun    bool
	keys      []string
}

// NewTruncater creates a new Truncater. maxLen must be > 0 and client/logger must be non-nil.
func NewTruncater(client *Client, logger *audit.Logger, maxLen int, keys []string, dryRun bool) (*Truncater, error) {
	if client == nil {
		return nil, fmt.Errorf("truncate: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("truncate: logger is required")
	}
	if maxLen <= 0 {
		return nil, fmt.Errorf("truncate: maxLen must be greater than zero")
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("truncate: at least one key must be specified")
	}
	return &Truncater{client: client, logger: logger, maxLen: maxLen, keys: keys, dryRun: dryRun}, nil
}

// Truncate reads the secret at path, truncates the specified keys, and writes back.
func (t *Truncater) Truncate(path string) error {
	secret, err := t.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("truncate: read %s: %w", path, err)
	}

	modified := false
	for _, k := range t.keys {
		v, ok := secret[k]
		if !ok {
			continue
		}
		str, ok := v.(string)
		if !ok {
			continue
		}
		if len(str) > t.maxLen {
			secret[k] = str[:t.maxLen]
			modified = true
		}
	}

	if !modified {
		t.logger.Log(map[string]interface{}{"op": "truncate", "path": path, "status": "skipped", "reason": "no values exceeded maxLen"})
		return nil
	}

	if t.dryRun {
		preview, _ := json.Marshal(secret)
		t.logger.Log(map[string]interface{}{"op": "truncate", "path": path, "status": "dry-run", "preview": string(preview)})
		return nil
	}

	if err := t.client.WriteSecret(path, secret); err != nil {
		return fmt.Errorf("truncate: write %s: %w", path, err)
	}

	t.logger.Log(map[string]interface{}{"op": "truncate", "path": path, "status": "ok"})
	return nil
}
