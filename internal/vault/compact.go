package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Compacter removes secrets whose data maps are empty (no keys or all keys
// have empty string values), keeping the engine tidy.
type Compacter struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
	removeEmpty bool // remove secrets with zero keys
	removeBlank bool // remove secrets where all values are ""
}

// NewCompacter creates a Compacter. client and logger are required.
func NewCompacter(client *Client, logger *audit.Logger, dryRun, removeEmpty, removeBlank bool) (*Compacter, error) {
	if client == nil {
		return nil, fmt.Errorf("compacter: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("compacter: logger is required")
	}
	if !removeEmpty && !removeBlank {
		return nil, fmt.Errorf("compacter: at least one of removeEmpty or removeBlank must be true")
	}
	return &Compacter{
		client:      client,
		logger:      logger,
		dryRun:      dryRun,
		removeEmpty: removeEmpty,
		removeBlank: removeBlank,
	}, nil
}

// Compact inspects each path and deletes those that qualify.
// It returns the list of paths deleted (or that would be deleted in dry-run).
func (c *Compacter) Compact(paths []string) ([]string, error) {
	var removed []string

	for _, p := range paths {
		data, err := c.client.ReadSecret(p)
		if err != nil {
			c.logger.Log("compact_read_error", map[string]interface{}{
				"path":  p,
				"error": err.Error(),
			})
			continue
		}

		if !c.shouldRemove(data) {
			continue
		}

		removed = append(removed, p)
		c.logger.Log("compact", map[string]interface{}{
			"path":    p,
			"dry_run": c.dryRun,
		})

		if c.dryRun {
			continue
		}

		if err := c.client.DeleteSecret(p); err != nil {
			c.logger.Log("compact_delete_error", map[string]interface{}{
				"path":  p,
				"error": err.Error(),
			})
		}
	}

	return removed, nil
}

func (c *Compacter) shouldRemove(data map[string]interface{}) bool {
	if c.removeEmpty && len(data) == 0 {
		return true
	}
	if c.removeBlank {
		for _, v := range data {
			if s, ok := v.(string); !ok || s != "" {
				return false
			}
		}
		if len(data) > 0 {
			return true
		}
	}
	return false
}
