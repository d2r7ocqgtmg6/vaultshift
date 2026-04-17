package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Redactor replaces secret values for specified keys with a placeholder.
type Redactor struct {
	client    *Client
	logger    *audit.Logger
	keys      []string
	placeholder string
}

// NewRedactor creates a Redactor. keys are the secret field names to redact.
func NewRedactor(client *Client, logger *audit.Logger, keys []string, placeholder string) (*Redactor, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one key is required")
	}
	if placeholder == "" {
		placeholder = "REDACTED"
	}
	return &Redactor{client: client, logger: logger, keys: keys, placeholder: placeholder}, nil
}

// Redact reads the secret at path, replaces matching keys, and writes it back.
func (r *Redactor) Redact(path string, dryRun bool) error {
	secret, err := r.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	data, ok := secret["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format at %s", path)
	}

	redacted := []string{}
	for _, k := range r.keys {
		if _, exists := data[k]; exists {
			data[k] = r.placeholder
			redacted = append(redacted, k)
		}
	}

	if len(redacted) == 0 {
		return nil
	}

	if r.logger != nil {
		r.logger.Log(map[string]interface{}{
			"action":   "redact",
			"path":     path,
			"keys":     strings.Join(redacted, ","),
			"dry_run":  dryRun,
		})
	}

	if dryRun {
		return nil
	}

	return r.client.WriteSecret(path, map[string]interface{}{"data": data})
}
