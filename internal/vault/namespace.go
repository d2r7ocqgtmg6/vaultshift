package vault

import (
	"fmt"

	"github.com/auditlog/logger"
)

// NamespaceMover moves all secrets from one namespace prefix to another.
type NamespaceMover struct {
	src    *Client
	dst    *Client
	logger *logger.Logger
	dryRun bool
}

// NewNamespaceMover creates a NamespaceMover. Both clients must be non-nil.
func NewNamespaceMover(src, dst *Client, log *logger.Logger, dryRun bool) (*NamespaceMover, error) {
	if src == nil {
		return nil, fmt.Errorf("namespace: source client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("namespace: destination client is required")
	}
	if log == nil {
		return nil, fmt.Errorf("namespace: logger is required")
	}
	return &NamespaceMover{src: src, dst: dst, logger: log, dryRun: dryRun}, nil
}

// Move copies all secrets under srcPrefix into dstPrefix.
func (n *NamespaceMover) Move(srcPrefix, dstPrefix string) (int, error) {
	paths, err := listAll(n.src, srcPrefix)
	if err != nil {
		return 0, fmt.Errorf("namespace move: list %q: %w", srcPrefix, err)
	}

	moved := 0
	for _, p := range paths {
		data, err := n.src.ReadSecret(p)
		if err != nil {
			n.logger.Log("namespace_move_read_error", map[string]interface{}{"path": p, "error": err.Error()})
			continue
		}

		rel := p[len(srcPrefix):]
		destPath := dstPrefix + rel

		if n.dryRun {
			n.logger.Log("namespace_move_dry_run", map[string]interface{}{"src": p, "dst": destPath})
			moved++
			continue
		}

		if err := n.dst.WriteSecret(destPath, data); err != nil {
			n.logger.Log("namespace_move_write_error", map[string]interface{}{"dst": destPath, "error": err.Error()})
			continue
		}

		n.logger.Log("namespace_moved", map[string]interface{}{"src": p, "dst": destPath})
		moved++
	}
	return moved, nil
}
