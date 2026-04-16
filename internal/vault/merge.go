package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// MergeResult holds the outcome of a single secret merge.
type MergeResult struct {
	Path      string
	Action    string // "created", "updated", "skipped"
	Conflict  bool
}

// Merger merges secrets from a source prefix into a destination prefix.
type Merger struct {
	src    *Client
	dst    *Client
	logger *audit.Logger
	dryRun bool
	overwrite bool
}

// NewMerger creates a new Merger.
func NewMerger(src, dst *Client, logger *audit.Logger, dryRun, overwrite bool) (*Merger, error) {
	if src == nil {
		return nil, fmt.Errorf("merge: source client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("merge: destination client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("merge: logger is required")
	}
	return &Merger{src: src, dst: dst, logger: logger, dryRun: dryRun, overwrite: overwrite}, nil
}

// Merge reads all secrets under srcPrefix and writes them under dstPrefix.
func (m *Merger) Merge(srcPrefix, dstPrefix string) ([]MergeResult, error) {
	paths, err := listAll(m.src, srcPrefix)
	if err != nil {
		return nil, fmt.Errorf("merge: list source: %w", err)
	}

	var results []MergeResult
	for _, p := range paths {
		rel := p[len(srcPrefix):]
		destPath := dstPrefix + rel

		data, err := m.src.ReadSecret(p)
		if err != nil {
			return results, fmt.Errorf("merge: read %s: %w", p, err)
		}

		existing, _ := m.dst.ReadSecret(destPath)
		action := "created"
		conflict := false
		if existing != nil {
			conflict = true
			if !m.overwrite {
				action = "skipped"
				m.logger.Log("merge", map[string]any{"path": destPath, "action": action, "dry_run": m.dryRun})
				results = append(results, MergeResult{Path: destPath, Action: action, Conflict: conflict})
				continue
			}
			action = "updated"
		}

		if !m.dryRun {
			if err := m.dst.WriteSecret(destPath, data); err != nil {
				return results, fmt.Errorf("merge: write %s: %w", destPath, err)
			}
		}
		m.logger.Log("merge", map[string]any{"path": destPath, "action": action, "dry_run": m.dryRun})
		results = append(results, MergeResult{Path: destPath, Action: action, Conflict: conflict})
	}
	return results, nil
}
