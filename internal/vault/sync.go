package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Syncer bidirectionally syncs secrets between two vault paths.
type Syncer struct {
	src    *Client
	dst    *Client
	logger *audit.Logger
	dryRun bool
}

type SyncResult struct {
	SrcToDst []string
	DstToSrc []string
	Errors   []string
}

func NewSyncer(src, dst *Client, logger *audit.Logger, dryRun bool) (*Syncer, error) {
	if src == nil {
		return nil, fmt.Errorf("src client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("dst client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Syncer{src: src, dst: dst, logger: logger, dryRun: dryRun}, nil
}

func (s *Syncer) Sync(srcPrefix, dstPrefix string) (*SyncResult, error) {
	result := &SyncResult{}

	srcPaths, err := listAll(s.src, srcPrefix)
	if err != nil {
		return nil, fmt.Errorf("listing src: %w", err)
	}
	dstPaths, err := listAll(s.dst, dstPrefix)
	if err != nil {
		return nil, fmt.Errorf("listing dst: %w", err)
	}

	dstSet := make(map[string]struct{}, len(dstPaths))
	for _, p := range dstPaths {
		dstSet[p] = struct{}{}
	}
	srcSet := make(map[string]struct{}, len(srcPaths))
	for _, p := range srcPaths {
		srcSet[p] = struct{}{}
	}

	for _, path := range srcPaths {
		if _, exists := dstSet[path]; !exists {
			s.logger.Log("sync", map[string]interface{}{"direction": "src->dst", "path": path, "dry_run": s.dryRun})
			if !s.dryRun {
				data, err := s.src.ReadSecret(srcPrefix + path)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("read src %s: %v", path, err))
					continue
				}
				if err := s.dst.WriteSecret(dstPrefix+path, data); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("write dst %s: %v", path, err))
					continue
				}
			}
			result.SrcToDst = append(result.SrcToDst, path)
		}
	}

	for _, path := range dstPaths {
		if _, exists := srcSet[path]; !exists {
			s.logger.Log("sync", map[string]interface{}{"direction": "dst->src", "path": path, "dry_run": s.dryRun})
			if !s.dryRun {
				data, err := s.dst.ReadSecret(dstPrefix + path)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("read dst %s: %v", path, err))
					continue
				}
				if err := s.src.WriteSecret(srcPrefix+path, data); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("write src %s: %v", path, err))
					continue
				}
			}
			result.DstToSrc = append(result.DstToSrc, path)
		}
	}

	return result, nil
}
