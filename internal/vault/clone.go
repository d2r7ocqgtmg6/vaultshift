package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// CloneResult holds the outcome of a clone operation.
type CloneResult struct {
	Copied  []string
	Skipped []string
	Errors  []string
}

// Cloner copies all secrets from one path prefix to another within the same client.
type Cloner struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewCloner creates a new Cloner instance.
func NewCloner(client *Client, logger *audit.Logger, dryRun bool) *Cloner {
	return &Cloner{
		client: client,
		logger: logger,
		dryRun: dryRun,
	}
}

// Clone copies all secrets under srcPrefix to destPrefix.
func (c *Cloner) Clone(srcPrefix, destPrefix string) (*CloneResult, error) {
	paths, err := listAll(c.client, srcPrefix)
	if err != nil {
		return nil, fmt.Errorf("clone: list %q: %w", srcPrefix, err)
	}

	result := &CloneResult{}

	for _, path := range paths {
		secret, err := c.client.ReadSecret(path)
		if err != nil {
			result.Errors = append(result.Errors, path)
			c.logger.Log("clone_read_error", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
			continue
		}

		relative := path[len(srcPrefix):]
		destPath := destPrefix + relative

		if c.dryRun {
			result.Skipped = append(result.Skipped, destPath)
			c.logger.Log("clone_dry_run", map[string]interface{}{
				"src":  path,
				"dest": destPath,
			})
			continue
		}

		if err := c.client.WriteSecret(destPath, secret); err != nil {
			result.Errors = append(result.Errors, destPath)
			c.logger.Log("clone_write_error", map[string]interface{}{
				"dest":  destPath,
				"error": err.Error(),
			})
			continue
		}

		result.Copied = append(result.Copied, destPath)
		c.logger.Log("clone_copied", map[string]interface{}{
			"src":  path,
			"dest": destPath,
		})
	}

	return result, nil
}
