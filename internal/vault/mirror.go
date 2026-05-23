package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// MirrorResult holds the outcome of a mirror operation.
type MirrorResult struct {
	Path    string
	Skipped bool
	Error   error
}

// Mirrorer continuously replicates secrets from a source prefix to a
// destination prefix, overwriting any existing values.
type Mirrorer struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewMirrorer constructs a Mirrorer. Returns an error if required fields are absent.
func NewMirrorer(client *Client, logger *audit.Logger, dryRun bool) (*Mirrorer, error) {
	if client == nil {
		return nil, fmt.Errorf("mirror: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("mirror: logger is required")
	}
	return &Mirrorer{client: client, logger: logger, dryRun: dryRun}, nil
}

// Mirror reads every secret under srcPrefix and writes it to the
// corresponding path under dstPrefix, replacing whatever was there.
func (m *Mirrorer) Mirror(srcPrefix, dstPrefix string) ([]MirrorResult, error) {
	paths, err := listAll(m.client, srcPrefix)
	if err != nil {
		return nil, fmt.Errorf("mirror: list %q: %w", srcPrefix, err)
	}

	var results []MirrorResult
	for _, rel := range paths {
		srcPath := srcPrefix + rel
		dstPath := dstPrefix + rel

		data, err := m.client.ReadSecret(srcPath)
		if err != nil {
			m.logger.Log("mirror", srcPath, "error", err.Error())
			results = append(results, MirrorResult{Path: srcPath, Error: err})
			continue
		}

		if m.dryRun {
			m.logger.Log("mirror", srcPath, "dry_run", "skipped write to "+dstPath)
			results = append(results, MirrorResult{Path: srcPath, Skipped: true})
			continue
		}

		if err := m.client.WriteSecret(dstPath, data); err != nil {
			m.logger.Log("mirror", srcPath, "error", err.Error())
			results = append(results, MirrorResult{Path: srcPath, Error: err})
			continue
		}

		m.logger.Log("mirror", srcPath, "status", "mirrored to "+dstPath)
		results = append(results, MirrorResult{Path: srcPath})
	}
	return results, nil
}
