package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Copier copies a single secret from one path to another within or across clients.
type Copier struct {
	src    *Client
	dest   *Client
	logger *audit.Logger
	dryRun bool
}

// NewCopier creates a Copier. src and dest may be the same client.
func NewCopier(src, dest *Client, logger *audit.Logger, dryRun bool) (*Copier, error) {
	if src == nil {
		return nil, fmt.Errorf("copy: source client is required")
	}
	if dest == nil {
		return nil, fmt.Errorf("copy: destination client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("copy: logger is required")
	}
	return &Copier{src: src, dest: dest, logger: logger, dryRun: dryRun}, nil
}

// Copy reads the secret at srcPath and writes it to destPath.
// In dry-run mode the write is skipped.
func (c *Copier) Copy(srcPath, destPath string) error {
	data, err := c.src.ReadSecret(srcPath)
	if err != nil {
		c.logger.Log("copy", srcPath, fmt.Sprintf("read error: %s", err))
		return fmt.Errorf("copy: read %q: %w", srcPath, err)
	}

	if c.dryRun {
		c.logger.Log("copy_dry_run", srcPath, fmt.Sprintf("would copy to %s", destPath))
		return nil
	}

	if err := c.dest.WriteSecret(destPath, data); err != nil {
		c.logger.Log("copy_error", destPath, fmt.Sprintf("write error: %s", err))
		return fmt.Errorf("copy: write %q: %w", destPath, err)
	}

	c.logger.Log("copy", srcPath, fmt.Sprintf("copied to %s", destPath))
	return nil
}
