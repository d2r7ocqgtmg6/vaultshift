package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Splitter reads secrets from a source prefix and writes them to multiple
// destination prefixes based on a key-based routing rule.
type Splitter struct {
	client  *api.Client
	logger  AuditLogger
	dryRun  bool
	routes  map[string]string // key suffix -> destination prefix
}

// SplitResult holds the outcome of a single secret split operation.
type SplitResult struct {
	SourcePath string
	DestPath   string
	Key        string
	Skipped    bool
}

// NewSplitter constructs a Splitter. routes maps a key name to a destination
// path prefix. Returns an error if client, logger, or routes are missing.
func NewSplitter(client *api.Client, logger AuditLogger, routes map[string]string, dryRun bool) (*Splitter, error) {
	if client == nil {
		return nil, fmt.Errorf("splitter: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("splitter: logger is required")
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("splitter: at least one route is required")
	}
	return &Splitter{client: client, logger: logger, dryRun: dryRun, routes: routes}, nil
}

// Split reads the secret at sourcePath and writes each routed key to its
// destination prefix. Keys not present in routes are ignored.
func (s *Splitter) Split(sourcePath string) ([]SplitResult, error) {
	secret, err := s.client.Logical().Read(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("splitter: read %s: %w", sourcePath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("splitter: secret not found at %s", sourcePath)
	}

	var results []SplitResult
	for key, destPrefix := range s.routes {
		val, ok := secret.Data[key]
		if !ok {
			continue
		}
		destPath := destPrefix + "/" + key
		result := SplitResult{SourcePath: sourcePath, DestPath: destPath, Key: key}
		if s.dryRun {
			result.Skipped = true
			s.logger.Log("split", sourcePath, "dry-run", nil)
			results = append(results, result)
			continue
		}
		_, err := s.client.Logical().Write(destPath, map[string]interface{}{"data": map[string]interface{}{key: val}})
		if err != nil {
			s.logger.Log("split", destPath, "error", err)
			return results, fmt.Errorf("splitter: write %s: %w", destPath, err)
		}
		s.logger.Log("split", destPath, "ok", nil)
		results = append(results, result)
	}
	return results, nil
}
