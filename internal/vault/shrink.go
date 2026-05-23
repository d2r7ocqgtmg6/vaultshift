package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Shrinker removes keys from secrets whose values exceed a maximum byte length.
type Shrinker struct {
	client  *Client
	logger  *audit.Logger
	maxLen  int
	dryRun  bool
}

// NewShrinker creates a Shrinker. maxLen is the maximum allowed value length in bytes.
func NewShrinker(client *Client, logger *audit.Logger, maxLen int, dryRun bool) (*Shrinker, error) {
	if client == nil {
		return nil, fmt.Errorf("shrink: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("shrink: logger is required")
	}
	if maxLen <= 0 {
		return nil, fmt.Errorf("shrink: maxLen must be greater than zero")
	}
	return &Shrinker{client: client, logger: logger, maxLen: maxLen, dryRun: dryRun}, nil
}

// Shrink reads the secret at path, removes any key whose value exceeds maxLen bytes,
// and writes the result back unless dryRun is enabled.
func (s *Shrinker) Shrink(path string) ([]string, error) {
	data, err := s.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("shrink: read %s: %w", path, err)
	}

	var removed []string
	cleaned := make(map[string]interface{}, len(data))
	for k, v := range data {
		str, ok := v.(string)
		if ok && len(str) > s.maxLen {
			removed = append(removed, k)
			s.logger.Log(map[string]interface{}{
				"op":     "shrink",
				"path":   path,
				"key":    k,
				"len":    len(str),
				"maxLen": s.maxLen,
				"dryRun": s.dryRun,
			})
			continue
		}
		cleaned[k] = v
	}

	if len(removed) == 0 {
		return nil, nil
	}

	if s.dryRun {
		return removed, nil
	}

	if err := s.client.WriteSecret(path, cleaned); err != nil {
		return nil, fmt.Errorf("shrink: write %s: %w", path, err)
	}
	return removed, nil
}
