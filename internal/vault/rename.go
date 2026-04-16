package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Renamer moves a secret from one path to another within the same client.
type Renamer struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewRenamer creates a Renamer. client and logger are required.
func NewRenamer(client *Client, logger *audit.Logger, dryRun bool) (*Renamer, error) {
	if client == nil {
		return nil, fmt.Errorf("rename: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("rename: logger is required")
	}
	return &Renamer{client: client, logger: logger, dryRun: dryRun}, nil
}

// Rename reads the secret at src, writes it to dst, then deletes src.
// In dry-run mode the write and delete are skipped.
func (r *Renamer) Rename(src, dst string) error {
	data, err := r.client.ReadSecret(src)
	if err != nil {
		return fmt.Errorf("rename: read %q: %w", src, err)
	}

	r.logger.Log("rename", map[string]any{
		"src":    src,
		"dst":    dst,
		"dryRun": r.dryRun,
	})

	if r.dryRun {
		return nil
	}

	if err := r.client.WriteSecret(dst, data); err != nil {
		return fmt.Errorf("rename: write %q: %w", dst, err)
	}

	if err := r.client.DeleteSecret(src); err != nil {
		return fmt.Errorf("rename: delete %q: %w", src, err)
	}

	return nil
}
