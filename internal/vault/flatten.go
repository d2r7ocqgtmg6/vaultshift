package vault

import (
	"fmt"

	"github.com/your-org/vaultshift/internal/audit"
)

// Flattener reads nested KV secrets under a prefix and writes them to a
// single destination path, merging all key-value pairs into one secret.
type Flattener struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
}

// NewFlattener constructs a Flattener. Returns an error if required fields
// are missing.
func NewFlattener(client *Client, logger *audit.Logger, dryRun bool) (*Flattener, error) {
	if client == nil {
		return nil, fmt.Errorf("flatten: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("flatten: logger is required")
	}
	return &Flattener{client: client, logger: logger, dryRun: dryRun}, nil
}

// Flatten lists all secrets under srcPrefix, merges their data into a single
// map, and writes the result to dstPath. In dry-run mode the write is skipped.
func (f *Flattener) Flatten(srcPrefix, dstPath string) error {
	paths, err := listAll(f.client, srcPrefix)
	if err != nil {
		return fmt.Errorf("flatten: list %q: %w", srcPrefix, err)
	}

	merged := make(map[string]interface{})
	for _, p := range paths {
		data, err := f.client.ReadSecret(p)
		if err != nil {
			f.logger.Log("flatten", map[string]interface{}{"path": p, "error": err.Error()})
			continue
		}
		for k, v := range data {
			merged[k] = v
		}
	}

	if len(merged) == 0 {
		f.logger.Log("flatten", map[string]interface{}{"src": srcPrefix, "result": "no data found"})
		return nil
	}

	if f.dryRun {
		f.logger.Log("flatten", map[string]interface{}{"dry_run": true, "dst": dstPath, "keys": len(merged)})
		return nil
	}

	if err := f.client.WriteSecret(dstPath, merged); err != nil {
		return fmt.Errorf("flatten: write %q: %w", dstPath, err)
	}

	f.logger.Log("flatten", map[string]interface{}{"src": srcPrefix, "dst": dstPath, "keys": len(merged)})
	return nil
}
