package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// ReplicateResult holds the outcome of a single secret replication.
type ReplicateResult struct {
	Path    string
	Skipped bool
	Error   error
}

// Replicator copies secrets from one prefix to multiple destination prefixes.
type Replicator struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
	overwrite bool
}

// NewReplicator constructs a Replicator.
func NewReplicator(client *Client, logger *audit.Logger, dryRun, overwrite bool) (*Replicator, error) {
	if client == nil {
		return nil, fmt.Errorf("replicate: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("replicate: logger is required")
	}
	return &Replicator{client: client, logger: logger, dryRun: dryRun, overwrite: overwrite}, nil
}

// Replicate reads all secrets under srcPrefix and writes them to each dest prefix.
func (r *Replicator) Replicate(srcPrefix string, destPrefixes []string) ([]ReplicateResult, error) {
	if len(destPrefixes) == 0 {
		return nil, fmt.Errorf("replicate: at least one destination prefix is required")
	}

	paths, err := listAll(r.client, srcPrefix)
	if err != nil {
		return nil, fmt.Errorf("replicate: list %q: %w", srcPrefix, err)
	}

	var results []ReplicateResult
	for _, path := range paths {
		data, err := r.client.ReadSecret(path)
		if err != nil {
			results = append(results, ReplicateResult{Path: path, Error: err})
			r.logger.Log("replicate", path, "error", err.Error())
			continue
		}

		rel := path[len(srcPrefix):]
		for _, dest := range destPrefixes {
			destPath := dest + rel
			if r.dryRun {
				r.logger.Log("replicate", destPath, "dry_run", "skipped")
				results = append(results, ReplicateResult{Path: destPath, Skipped: true})
				continue
			}
			if !r.overwrite {
				existing, _ := r.client.ReadSecret(destPath)
				if existing != nil {
					r.logger.Log("replicate", destPath, "status", "exists_skipped")
					results = append(results, ReplicateResult{Path: destPath, Skipped: true})
					continue
				}
			}
			if wErr := r.client.WriteSecret(destPath, data); wErr != nil {
				r.logger.Log("replicate", destPath, "error", wErr.Error())
				results = append(results, ReplicateResult{Path: destPath, Error: wErr})
				continue
			}
			r.logger.Log("replicate", destPath, "status", "written")
			results = append(results, ReplicateResult{Path: destPath})
		}
	}
	return results, nil
}
