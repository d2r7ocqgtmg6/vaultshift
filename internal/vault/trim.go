package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Trimmer removes leading/trailing whitespace from secret values.
type Trimmer struct {
	client *Client
	logger *audit.Logger
	dryRun bool
	keys   []string // if empty, trim all string values
}

// NewTrimmer creates a Trimmer. client and logger are required.
func NewTrimmer(client *Client, logger *audit.Logger, dryRun bool, keys []string) (*Trimmer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Trimmer{client: client, logger: logger, dryRun: dryRun, keys: keys}, nil
}

// Trim reads the secret at path, trims whitespace from targeted keys, and writes it back.
func (t *Trimmer) Trim(path string) (bool, error) {
	data, err := t.client.ReadSecret(path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	modified := false
	for k, v := range data {
		if !t.shouldTrim(k) {
			continue
		}
		str, ok := v.(string)
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(str)
		if trimmed != str {
			data[k] = trimmed
			modified = true
		}
	}

	if !modified {
		t.logger.Log("trim", path, "no changes", nil)
		return false, nil
	}

	if t.dryRun {
		t.logger.Log("trim", path, "dry-run", nil)
		return true, nil
	}

	if err := t.client.WriteSecret(path, data); err != nil {
		return false, fmt.Errorf("write %s: %w", path, err)
	}
	t.logger.Log("trim", path, "trimmed", nil)
	return true, nil
}

func (t *Trimmer) shouldTrim(key string) bool {
	if len(t.keys) == 0 {
		return true
	}
	for _, k := range t.keys {
		if k == key {
			return true
		}
	}
	return false
}
