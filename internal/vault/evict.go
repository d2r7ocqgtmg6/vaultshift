package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Evicter removes secrets whose keys match a given set of patterns.
type Evicter struct {
	client  *Client
	logger  *audit.Logger
	patterns []string
	dryRun  bool
}

// EvictResult holds the outcome of an eviction operation.
type EvictResult struct {
	Path    string
	Key     string
	Evicted bool
}

// NewEvicter constructs an Evicter. Returns an error if required fields are missing.
func NewEvicter(client *Client, logger *audit.Logger, patterns []string, dryRun bool) (*Evicter, error) {
	if client == nil {
		return nil, fmt.Errorf("evict: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("evict: logger is required")
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("evict: at least one pattern is required")
	}
	return &Evicter{
		client:   client,
		logger:   logger,
		patterns: patterns,
		dryRun:   dryRun,
	}, nil
}

// Evict reads the secret at path, removes matching keys, and writes back the result.
// Returns a slice of EvictResult describing which keys were evicted.
func (e *Evicter) Evict(path string) ([]EvictResult, error) {
	data, err := e.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("evict: read %s: %w", path, err)
	}

	var results []EvictResult
	updated := make(map[string]interface{})

	for k, v := range data {
		if e.matchesAny(k) {
			results = append(results, EvictResult{Path: path, Key: k, Evicted: true})
			e.logger.Log("evict", map[string]interface{}{
				"path":    path,
				"key":     k,
				"dry_run": e.dryRun,
			})
		} else {
			updated[k] = v
			results = append(results, EvictResult{Path: path, Key: k, Evicted: false})
		}
	}

	if e.dryRun {
		return results, nil
	}

	if err := e.client.WriteSecret(path, updated); err != nil {
		return nil, fmt.Errorf("evict: write %s: %w", path, err)
	}

	return results, nil
}

func (e *Evicter) matchesAny(key string) bool {
	for _, p := range e.patterns {
		if strings.Contains(key, p) {
			return true
		}
	}
	return false
}
