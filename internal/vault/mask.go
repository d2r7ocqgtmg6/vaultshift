package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Masker replaces secret values with a masked placeholder.
type Masker struct {
	client  *Client
	logger  *audit.Logger
	keys    []string
	mask    string
	dryRun  bool
}

// NewMasker creates a Masker that will replace specified keys with a mask string.
func NewMasker(client *Client, logger *audit.Logger, keys []string, mask string, dryRun bool) (*Masker, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one key is required")
	}
	if mask == "" {
		mask = "***"
	}
	return &Masker{client: client, logger: logger, keys: keys, mask: mask, dryRun: dryRun}, nil
}

// Mask reads the secret at path, replaces matching keys with the mask value, and writes it back.
func (m *Masker) Mask(path string) error {
	secret, err := m.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	keySet := make(map[string]bool, len(m.keys))
	for _, k := range m.keys {
		keySet[strings.TrimSpace(k)] = true
	}

	masked := make(map[string]interface{}, len(secret))
	for k, v := range secret {
		if keySet[k] {
			masked[k] = m.mask
		} else {
			masked[k] = v
		}
	}

	m.logger.Log(map[string]interface{}{
		"op":     "mask",
		"path":   path,
		"keys":   m.keys,
		"dryRun": m.dryRun,
	})

	if m.dryRun {
		return nil
	}

	if err := m.client.WriteSecret(path, masked); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
