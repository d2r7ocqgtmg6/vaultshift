package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Deduper detects duplicate secret values across paths.
type Deduper struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// DedupeResult holds a group of paths that share the same secret data.
type DedupeResult struct {
	Paths  []string
	Sample map[string]interface{}
}

// NewDeduper creates a new Deduper.
func NewDeduper(client *Client, logger *audit.Logger, dryRun bool) (*Deduper, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Deduper{client: client, logger: logger, dryRun: dryRun}, nil
}

// Dedupe scans secrets under prefix and returns groups of duplicate paths.
func (d *Deduper) Dedupe(prefix string) ([]DedupeResult, error) {
	lister := &lister{client: d.client}
	paths, err := lister.listAll(prefix)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	type entry struct {
		key  string
		data map[string]interface{}
	}

	seen := map[string][]string{}
	samples := map[string]map[string]interface{}{}

	for _, path := range paths {
		data, err := d.client.ReadSecret(path)
		if err != nil {
			d.logger.Log("dedupe_read_error", map[string]interface{}{"path": path, "error": err.Error()})
			continue
		}
		key := canonicalKey(data)
		seen[key] = append(seen[key], path)
		if _, ok := samples[key]; !ok {
			samples[key] = data
		}
	}

	var results []DedupeResult
	for key, paths := range seen {
		if len(paths) > 1 {
			results = append(results, DedupeResult{Paths: paths, Sample: samples[key]})
			d.logger.Log("dedupe_found", map[string]interface{}{"paths": paths, "dry_run": d.dryRun})
		}
	}
	return results, nil
}

// canonicalKey builds a stable string key from secret data for comparison.
func canonicalKey(data map[string]interface{}) string {
	h := fmt.Sprintf("%v", data)
	return h
}
