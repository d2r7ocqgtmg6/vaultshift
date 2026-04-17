package vault

import (
	"fmt"
	"strings"

	"github.com/your-org/vaultshift/internal/audit"
)

// Sanitizer removes or redacts secret keys matching configured patterns.
type Sanitizer struct {
	client  *Client
	logger  *audit.Logger
	keys    []string
	dryRun  bool
}

type SanitizerConfig struct {
	Client  *Client
	Logger  *audit.Logger
	Keys    []string
	DryRun  bool
}

func NewSanitizer(cfg SanitizerConfig) (*Sanitizer, error) {
	if cfg.Client == nil {
		return nil, fmt.Errorf("sanitizer: client is required")
	}
	if len(cfg.Keys) == 0 {
		return nil, fmt.Errorf("sanitizer: at least one key pattern is required")
	}
	return &Sanitizer{
		client: cfg.Client,
		logger: cfg.Logger,
		keys:   cfg.Keys,
		dryRun: cfg.DryRun,
	}, nil
}

// Sanitize reads the secret at path, removes matching keys, and writes it back.
func (s *Sanitizer) Sanitize(path string) (bool, error) {
	data, err := s.client.ReadSecret(path)
	if err != nil {
		return false, fmt.Errorf("sanitize: read %s: %w", path, err)
	}

	modified := false
	for k := range data {
		if s.matchesAny(k) {
			if s.logger != nil {
				s.logger.Log("sanitize", map[string]any{"path": path, "key": k, "dry_run": s.dryRun})
			}
			if !s.dryRun {
				delete(data, k)
			}
			modified = true
		}
	}

	if modified && !s.dryRun {
		if err := s.client.WriteSecret(path, data); err != nil {
			return false, fmt.Errorf("sanitize: write %s: %w", path, err)
		}
	}

	return modified, nil
}

func (s *Sanitizer) matchesAny(key string) bool {
	for _, pattern := range s.keys {
		if strings.EqualFold(key, pattern) || strings.HasPrefix(strings.ToLower(key), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
