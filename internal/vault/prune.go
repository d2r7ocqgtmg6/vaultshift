package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Pruner removes secrets from a path that are not present in a reference set.
type Pruner struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewPruner creates a Pruner. Returns an error if client or logger is nil.
func NewPruner(client *Client, logger *audit.Logger, dryRun bool) (*Pruner, error) {
	if client == nil {
		return nil, fmt.Errorf("prune: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("prune: logger is required")
	}
	return &Pruner{client: client, logger: logger, dryRun: dryRun}, nil
}

// PruneResult holds the outcome of a prune operation.
type PruneResult struct {
	Deleted []string
	Skipped []string
	Errors  []string
}

// Prune deletes secrets under destPrefix that are absent from keepPaths.
func (p *Pruner) Prune(destPrefix string, keepPaths []string) (*PruneResult, error) {
	keepSet := make(map[string]struct{}, len(keepPaths))
	for _, k := range keepPaths {
		keepSet[k] = struct{}{}
	}

	existing, err := listAll(p.client, destPrefix)
	if err != nil {
		return nil, fmt.Errorf("prune: list %q: %w", destPrefix, err)
	}

	result := &PruneResult{}
	for _, path := range existing {
		if _, keep := keepSet[path]; keep {
			result.Skipped = append(result.Skipped, path)
			continue
		}
		if p.dryRun {
			p.logger.Log("prune", path, "dry-run", nil)
			result.Skipped = append(result.Skipped, path)
			continue
		}
		if err := p.client.DeleteSecret(path); err != nil {
			p.logger.Log("prune", path, "error", err)
			result.Errors = append(result.Errors, path)
			continue
		}
		p.logger.Log("prune", path, "deleted", nil)
		result.Deleted = append(result.Deleted, path)
	}
	return result, nil
}
