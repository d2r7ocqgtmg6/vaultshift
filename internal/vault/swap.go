package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Swapper exchanges the values of two secrets at different paths.
type Swapper struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewSwapper creates a Swapper. Returns an error if client or logger is nil.
func NewSwapper(client *Client, logger *audit.Logger, dryRun bool) (*Swapper, error) {
	if client == nil {
		return nil, fmt.Errorf("swap: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("swap: logger is required")
	}
	return &Swapper{client: client, logger: logger, dryRun: dryRun}, nil
}

// Swap reads the secrets at pathA and pathB, then writes each value to the
// other path. In dry-run mode no writes are performed.
func (s *Swapper) Swap(pathA, pathB string) error {
	dataA, err := s.client.ReadSecret(pathA)
	if err != nil {
		return fmt.Errorf("swap: read %s: %w", pathA, err)
	}
	dataB, err := s.client.ReadSecret(pathB)
	if err != nil {
		return fmt.Errorf("swap: read %s: %w", pathB, err)
	}

	if s.dryRun {
		s.logger.Log("swap", map[string]interface{}{
			"dry_run": true,
			"path_a":  pathA,
			"path_b":  pathB,
		})
		return nil
	}

	if err := s.client.WriteSecret(pathA, dataB); err != nil {
		return fmt.Errorf("swap: write %s: %w", pathA, err)
	}
	if err := s.client.WriteSecret(pathB, dataA); err != nil {
		return fmt.Errorf("swap: write %s: %w", pathB, err)
	}

	s.logger.Log("swap", map[string]interface{}{
		"path_a": pathA,
		"path_b": pathB,
		"status": "swapped",
	})
	return nil
}
