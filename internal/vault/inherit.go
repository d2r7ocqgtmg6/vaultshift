package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Inheriter copies secrets from a parent path into a child path,
// skipping keys that already exist in the child (unless overwrite is set).
type Inheriter struct {
	client    *Client
	logger    *audit.Logger
	dryRun    bool
	overwrite bool
}

func NewInheritor(client *Client, logger *audit.Logger, dryRun, overwrite bool) (*Inheriter, error) {
	if client == nil {
		return nil, fmt.Errorf("inherit: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("inherit: logger is required")
	}
	return &Inheriter{client: client, logger: logger, dryRun: dryRun, overwrite: overwrite}, nil
}

// Inherit reads secrets from parentPath and writes missing keys into childPath.
func (h *Inheriter) Inherit(parentPath, childPath string) (int, error) {
	parent, err := h.client.ReadSecret(parentPath)
	if err != nil {
		return 0, fmt.Errorf("inherit: read parent %q: %w", parentPath, err)
	}

	child, err := h.client.ReadSecret(childPath)
	if err != nil {
		child = map[string]interface{}{}
	}

	merged := make(map[string]interface{})
	for k, v := range child {
		merged[k] = v
	}

	inherited := 0
	for k, v := range parent {
		if _, exists := merged[k]; exists && !h.overwrite {
			continue
		}
		merged[k] = v
		inherited++
	}

	if inherited == 0 {
		h.logger.Log("inherit", map[string]interface{}{"parent": parentPath, "child": childPath, "inherited": 0})
		return 0, nil
	}

	if h.dryRun {
		h.logger.Log("inherit_dry_run", map[string]interface{}{"parent": parentPath, "child": childPath, "would_inherit": inherited})
		return inherited, nil
	}

	if err := h.client.WriteSecret(childPath, merged); err != nil {
		return 0, fmt.Errorf("inherit: write child %q: %w", childPath, err)
	}

	h.logger.Log("inherited", map[string]interface{}{"parent": parentPath, "child": childPath, "inherited": inherited})
	return inherited, nil
}
