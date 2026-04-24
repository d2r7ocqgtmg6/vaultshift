package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Normalizer rewrites secret values by applying string normalization rules
// such as trimming whitespace, lowercasing keys, or collapsing repeated delimiters.
type Normalizer struct {
	client    *Client
	logger    *audit.Logger
	trimSpace bool
	lowerKeys bool
	dryRun    bool
}

// NormalizerOption configures a Normalizer.
type NormalizerOption func(*Normalizer)

// WithNormalizeTrimSpace enables trimming of leading/trailing whitespace from values.
func WithNormalizeTrimSpace() NormalizerOption {
	return func(n *Normalizer) { n.trimSpace = true }
}

// WithNormalizeLowerKeys enables lowercasing of all secret keys.
func WithNormalizeLowerKeys() NormalizerOption {
	return func(n *Normalizer) { n.lowerKeys = true }
}

// WithNormalizeDryRun enables dry-run mode (no writes performed).
func WithNormalizeDryRun() NormalizerOption {
	return func(n *Normalizer) { n.dryRun = true }
}

// NewNormalizer constructs a Normalizer with the given options.
func NewNormalizer(client *Client, logger *audit.Logger, opts ...NormalizerOption) (*Normalizer, error) {
	if client == nil {
		return nil, fmt.Errorf("normalize: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("normalize: logger is required")
	}
	n := &Normalizer{client: client, logger: logger}
	for _, o := range opts {
		o(n)
	}
	return n, nil
}

// Normalize reads the secret at path, applies normalization rules, and writes
// the result back. In dry-run mode the transformed data is logged but not written.
func (n *Normalizer) Normalize(path string) error {
	data, err := n.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("normalize: read %q: %w", path, err)
	}

	normalized := make(map[string]interface{}, len(data))
	for k, v := range data {
		newKey := k
		if n.lowerKeys {
			newKey = strings.ToLower(k)
		}
		newVal := v
		if n.trimSpace {
			if s, ok := v.(string); ok {
				newVal = strings.TrimSpace(s)
			}
		}
		normalized[newKey] = newVal
	}

	if n.dryRun {
		n.logger.Log("normalize", map[string]interface{}{
			"dry_run": true,
			"path":    path,
			"keys":    len(normalized),
		})
		return nil
	}

	if err := n.client.WriteSecret(path, normalized); err != nil {
		return fmt.Errorf("normalize: write %q: %w", path, err)
	}

	n.logger.Log("normalize", map[string]interface{}{
		"path": path,
		"keys": len(normalized),
	})
	return nil
}
