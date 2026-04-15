package vault

import (
	"fmt"
	"strings"
)

// PromoteResult holds the outcome of a single secret promotion.
type PromoteResult struct {
	Path    string
	Skipped bool
	DryRun  bool
	Err     error
}

// Promoter moves secrets from a staging prefix to a production prefix.
type Promoter struct {
	src    *Client
	dst    *Client
	logger AuditLogger
	dryRun bool
}

// NewPromoter creates a Promoter that reads from src and writes to dst.
func NewPromoter(src, dst *Client, logger AuditLogger, dryRun bool) *Promoter {
	return &Promoter{src: src, dst: dst, logger: logger, dryRun: dryRun}
}

// Promote copies all secrets under srcPrefix into dstPrefix.
// Existing destination secrets are overwritten unless skipExisting is true.
func (p *Promoter) Promote(srcPrefix, dstPrefix string, skipExisting bool) ([]PromoteResult, error) {
	paths, err := listAll(p.src, srcPrefix)
	if err != nil {
		return nil, fmt.Errorf("promote: list %q: %w", srcPrefix, err)
	}

	var results []PromoteResult
	for _, path := range paths {
		rel := strings.TrimPrefix(path, srcPrefix)
		destPath := dstPrefix + rel

		result := PromoteResult{Path: destPath, DryRun: p.dryRun}

		if skipExisting {
			existing, _ := p.dst.ReadSecret(destPath)
			if existing != nil {
				result.Skipped = true
				p.logger.Log("promote", map[string]interface{}{"path": destPath, "status": "skipped"})
				results = append(results, result)
				continue
			}
		}

		data, err := p.src.ReadSecret(path)
		if err != nil {
			result.Err = fmt.Errorf("read %q: %w", path, err)
			p.logger.Log("promote", map[string]interface{}{"path": destPath, "status": "error", "error": result.Err.Error()})
			results = append(results, result)
			continue
		}

		if !p.dryRun {
			if err := p.dst.WriteSecret(destPath, data); err != nil {
				result.Err = fmt.Errorf("write %q: %w", destPath, err)
				p.logger.Log("promote", map[string]interface{}{"path": destPath, "status": "error", "error": result.Err.Error()})
				results = append(results, result)
				continue
			}
		}

		p.logger.Log("promote", map[string]interface{}{"path": destPath, "status": "ok", "dry_run": p.dryRun})
		results = append(results, result)
	}
	return results, nil
}
