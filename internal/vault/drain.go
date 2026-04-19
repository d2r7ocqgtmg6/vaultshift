package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Drainer moves all secrets from a source prefix then deletes them.
type Drainer struct {
	src    *Client
	dst    *Client
	logger *audit.Logger
	dryRun bool
}

// DrainOption configures a Drainer.
type DrainOption func(*Drainer)

// WithDrainDryRun enables dry-run mode.
func WithDrainDryRun(d bool) DrainOption {
	return func(dr *Drainer) { dr.dryRun = d }
}

// NewDrainer constructs a Drainer.
func NewDrainer(src, dst *Client, logger *audit.Logger, opts ...DrainOption) (*Drainer, error) {
	if src == nil {
		return nil, fmt.Errorf("drain: source client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("drain: destination client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("drain: logger is required")
	}
	dr := &Drainer{src: src, dst: dst, logger: logger}
	for _, o := range opts {
		o(dr)
	}
	return dr, nil
}

// Drain copies secrets from srcPrefix to dstPrefix then deletes the originals.
func (dr *Drainer) Drain(srcPrefix, dstPrefix string) (int, error) {
	paths, err := listAll(dr.src, srcPrefix)
	if err != nil {
		return 0, fmt.Errorf("drain: list %s: %w", srcPrefix, err)
	}

	moved := 0
	for _, p := range paths {
		data, err := dr.src.ReadSecret(p)
		if err != nil {
			dr.logger.Log("drain_read_error", map[string]interface{}{"path": p, "error": err.Error()})
			continue
		}

		rel := p[len(srcPrefix):]
		dest := dstPrefix + rel

		if dr.dryRun {
			dr.logger.Log("drain_dry_run", map[string]interface{}{"src": p, "dst": dest})
			moved++
			continue
		}

		if err := dr.dst.WriteSecret(dest, data); err != nil {
			dr.logger.Log("drain_write_error", map[string]interface{}{"path": dest, "error": err.Error()})
			continue
		}
		if err := dr.src.DeleteSecret(p); err != nil {
			dr.logger.Log("drain_delete_error", map[string]interface{}{"path": p, "error": err.Error()})
			continue
		}
		dr.logger.Log("drain_moved", map[string]interface{}{"src": p, "dst": dest})
		moved++
	}
	return moved, nil
}
