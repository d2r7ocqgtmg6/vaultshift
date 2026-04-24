package vault

import (
	"fmt"

	"github.com/dreamsofcode-io/vaultshift/internal/audit"
)

// Pivoter reorganises secrets by swapping the key namespace prefix between
// two paths, effectively "pivoting" a set of secrets from one mount point
// to another while preserving key names.
type Pivoter struct {
	src    *Client
	dst    *Client
	logger *audit.Logger
	dryRun bool
}

// NewPivoter creates a Pivoter. Both src and dst must be non-nil.
func NewPivoter(src, dst *Client, logger *audit.Logger, dryRun bool) (*Pivoter, error) {
	if src == nil {
		return nil, fmt.Errorf("pivot: source client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("pivot: destination client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("pivot: logger is required")
	}
	return &Pivoter{src: src, dst: dst, logger: logger, dryRun: dryRun}, nil
}

// Pivot reads every path in paths from src and writes it to dst under the
// same relative path. In dry-run mode no writes are performed.
func (p *Pivoter) Pivot(paths []string) error {
	var errs []string
	for _, path := range paths {
		data, err := p.src.ReadSecret(path)
		if err != nil {
			p.logger.Log("pivot_read_error", map[string]interface{}{"path": path, "error": err.Error()})
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		if p.dryRun {
			p.logger.Log("pivot_dry_run", map[string]interface{}{"path": path})
			continue
		}

		if err := p.dst.WriteSecret(path, data); err != nil {
			p.logger.Log("pivot_write_error", map[string]interface{}{"path": path, "error": err.Error()})
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		p.logger.Log("pivot_ok", map[string]interface{}{"path": path})
	}

	if len(errs) > 0 {
		return fmt.Errorf("pivot: %d error(s): %v", len(errs), errs)
	}
	return nil
}
