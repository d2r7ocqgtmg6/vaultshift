package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Cascader propagates a set of key-value pairs from a source secret to one or
// more destination paths, optionally skipping existing keys.
type Cascader struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
	overwrite bool
}

// CascadeResult holds the outcome for a single destination path.
type CascadeResult struct {
	Path    string
	Written bool
	Skipped int
	Err     error
}

// NewCascader constructs a Cascader. client and logger are required.
func NewCascader(client *Client, logger *audit.Logger, dryRun, overwrite bool) (*Cascader, error) {
	if client == nil {
		return nil, fmt.Errorf("cascade: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("cascade: logger is required")
	}
	return &Cascader{
		client:    client,
		logger:    logger,
		dryRun:    dryRun,
		overwrite: overwrite,
	}, nil
}

// Cascade reads the secret at srcPath and writes selected keys to each
// destination path. If overwrite is false, existing keys at the destination
// are preserved.
func (c *Cascader) Cascade(srcPath string, dstPaths []string, keys []string) ([]CascadeResult, error) {
	srcData, err := c.client.ReadSecret(srcPath)
	if err != nil {
		return nil, fmt.Errorf("cascade: read source %q: %w", srcPath, err)
	}

	// Build the subset of data to cascade.
	payload := make(map[string]interface{})
	if len(keys) == 0 {
		for k, v := range srcData {
			payload[k] = v
		}
	} else {
		for _, k := range keys {
			if v, ok := srcData[k]; ok {
				payload[k] = v
			}
		}
	}

	var results []CascadeResult
	for _, dst := range dstPaths {
		res := c.cascadeTo(dst, payload)
		results = append(results, res)
	}
	return results, nil
}

func (c *Cascader) cascadeTo(dst string, payload map[string]interface{}) CascadeResult {
	result := CascadeResult{Path: dst}

	existing, _ := c.client.ReadSecret(dst)
	merged := make(map[string]interface{})
	skipped := 0

	for k, v := range payload {
		if !c.overwrite {
			if _, exists := existing[k]; exists {
				skipped++
				continue
			}
		}
		merged[k] = v
	}

	// Carry over keys from existing that are not in payload.
	for k, v := range existing {
		if _, inPayload := payload[k]; !inPayload {
			merged[k] = v
		}
	}

	result.Skipped = skipped

	if c.dryRun {
		c.logger.Log("cascade", map[string]interface{}{"dry_run": true, "dst": dst, "skipped": skipped})
		result.Written = false
		return result
	}

	if err := c.client.WriteSecret(dst, merged); err != nil {
		c.logger.Log("cascade_error", map[string]interface{}{"dst": dst, "error": err.Error()})
		result.Err = err
		return result
	}

	c.logger.Log("cascade", map[string]interface{}{"dst": dst, "skipped": skipped})
	result.Written = true
	return result
}
