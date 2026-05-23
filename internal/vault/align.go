package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Aligner ensures that a set of secret paths in a destination client
// matches a reference set from a source client, writing missing entries
// and optionally removing orphaned ones.
type Aligner struct {
	src    *Client
	dst    *Client
	logger *audit.Logger
	dryRun bool
	prune  bool
}

// AlignResult holds the outcome of an alignment operation.
type AlignResult struct {
	Written  []string
	Pruned   []string
	Errors   []string
}

// NewAligner constructs an Aligner. src and dst must be non-nil.
func NewAligner(src, dst *Client, logger *audit.Logger, dryRun, prune bool) (*Aligner, error) {
	if src == nil {
		return nil, fmt.Errorf("align: source client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("align: destination client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("align: logger is required")
	}
	return &Aligner{src: src, dst: dst, logger: logger, dryRun: dryRun, prune: prune}, nil
}

// Align synchronises dst to match src under the given prefix.
func (a *Aligner) Align(prefix string) (*AlignResult, error) {
	srcPaths, err := listAll(a.src, prefix)
	if err != nil {
		return nil, fmt.Errorf("align: list source: %w", err)
	}
	dstPaths, err := listAll(a.dst, prefix)
	if err != nil {
		return nil, fmt.Errorf("align: list dest: %w", err)
	}

	srcSet := make(map[string]struct{}, len(srcPaths))
	for _, p := range srcPaths {
		srcSet[p] = struct{}{}
	}
	dstSet := make(map[string]struct{}, len(dstPaths))
	for _, p := range dstPaths {
		dstSet[p] = struct{}{}
	}

	result := &AlignResult{}

	// Write paths present in src but missing in dst.
	for _, p := range srcPaths {
		if _, ok := dstSet[p]; ok {
			continue
		}
		data, err := a.src.ReadSecret(p)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("read %s: %v", p, err))
			continue
		}
		if a.dryRun {
			a.logger.Log("align", p, "dry-run write skipped", nil)
			result.Written = append(result.Written, p)
			continue
		}
		if err := a.dst.WriteSecret(p, data); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("write %s: %v", p, err))
			continue
		}
		a.logger.Log("align", p, "written", nil)
		result.Written = append(result.Written, p)
	}

	// Prune paths present in dst but absent in src.
	if a.prune {
		for _, p := range dstPaths {
			if _, ok := srcSet[p]; ok {
				continue
			}
			if a.dryRun {
				a.logger.Log("align", p, "dry-run prune skipped", nil)
				result.Pruned = append(result.Pruned, p)
				continue
			}
			if err := a.dst.DeleteSecret(p); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("delete %s: %v", p, err))
				continue
			}
			a.logger.Log("align", p, "pruned", nil)
			result.Pruned = append(result.Pruned, p)
		}
	}

	return result, nil
}
