package vault

import (
	"errors"
	"fmt"
	"strings"

	"github.com/your-org/vaultshift/internal/audit"
)

// Tokenizer replaces sensitive secret values with opaque tokens stored at a target path.
type Tokenizer struct {
	client  *Client
	logger  *audit.Logger
	tokenNS string
	dryRun  bool
}

// NewTokenizer creates a Tokenizer. tokenNS is the KV prefix where tokens are stored.
func NewTokenizer(client *Client, logger *audit.Logger, tokenNS string, dryRun bool) (*Tokenizer, error) {
	if client == nil {
		return nil, errors.New("tokenizer: client is required")
	}
	if logger == nil {
		return nil, errors.New("tokenizer: logger is required")
	}
	if strings.TrimSpace(tokenNS) == "" {
		return nil, errors.New("tokenizer: tokenNS is required")
	}
	return &Tokenizer{client: client, logger: logger, tokenNS: tokenNS, dryRun: dryRun}, nil
}

// Tokenize reads the secret at path, replaces values for the given keys with
// token references, and writes the token mapping under tokenNS.
func (t *Tokenizer) Tokenize(path string, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return nil, errors.New("tokenize: at least one key is required")
	}

	secret, err := t.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("tokenize: read %s: %w", path, err)
	}

	tokens := make(map[string]string, len(keys))
	updated := make(map[string]interface{}, len(secret))
	for k, v := range secret {
		updated[k] = v
	}

	for _, key := range keys {
		val, ok := secret[key]
		if !ok {
			continue
		}
		token := fmt.Sprintf("tok:%s:%s", sanitizePath(path), key)
		tokens[key] = token
		updated[key] = token

		if !t.dryRun {
			tokenPath := fmt.Sprintf("%s/%s", strings.TrimRight(t.tokenNS, "/"), token)
			if err := t.client.WriteSecret(tokenPath, map[string]interface{}{"value": val}); err != nil {
				return nil, fmt.Errorf("tokenize: write token %s: %w", tokenPath, err)
			}
		}
	}

	if !t.dryRun {
		if err := t.client.WriteSecret(path, updated); err != nil {
			return nil, fmt.Errorf("tokenize: update secret %s: %w", path, err)
		}
	}

	t.logger.Log("tokenize", map[string]interface{}{
		"path":    path,
		"keys":    keys,
		"dry_run": t.dryRun,
	})
	return tokens, nil
}

func sanitizePath(p string) string {
	return strings.NewReplacer("/", "_", " ", "_").Replace(p)
}
