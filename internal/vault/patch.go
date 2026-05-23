package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Patcher applies partial updates to secrets at a given path,
// merging provided keys into the existing secret data.
type Patcher struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// PatchResult holds the outcome of a single patch operation.
type PatchResult struct {
	Path    string
	Applied map[string]string
	Skipped []string
	Err     error
}

// NewPatcher creates a Patcher. Returns an error if client or logger is nil.
func NewPatcher(client *Client, logger *audit.Logger) (*Patcher, error) {
	if client == nil {
		return nil, fmt.Errorf("patch: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("patch: logger is required")
	}
	return &Patcher{client: client, logger: logger}, nil
}

// WithPatchDryRun enables dry-run mode.
func (p *Patcher) WithPatchDryRun(v bool) *Patcher {
	p.dryRun = v
	return p
}

// Patch merges updates into the secret at path. Keys in updates overwrite
// existing values; keys absent from updates are preserved.
func (p *Patcher) Patch(path string, updates map[string]string) PatchResult {
	result := PatchResult{Path: path, Applied: map[string]string{}}

	existing, err := p.client.ReadSecret(path)
	if err != nil {
		result.Err = fmt.Errorf("patch: read %s: %w", path, err)
		p.logger.Log("patch", path, "error", result.Err.Error())
		return result
	}

	merged := make(map[string]interface{})
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range updates {
		merged[k] = v
		result.Applied[k] = v
	}

	if p.dryRun {
		p.logger.Log("patch", path, "dry_run", "true")
		return result
	}

	if err := p.client.WriteSecret(path, merged); err != nil {
		result.Err = fmt.Errorf("patch: write %s: %w", path, err)
		p.logger.Log("patch", path, "error", result.Err.Error())
		return result
	}

	p.logger.Log("patch", path, "status", "ok")
	return result
}
