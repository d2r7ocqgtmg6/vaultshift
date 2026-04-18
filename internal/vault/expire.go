package vault

import (
	"fmt"
	"time"

	"github.com/yourusername/vaultshift/internal/audit"
)

// ExpiryResult holds the outcome of an expiry check or purge for a single secret.
type ExpiryResult struct {
	Path    string
	Expired bool
	Error   error
}

// Expirer checks secrets for a metadata key (e.g. "expires_at") and optionally
// deletes those that have passed their expiry timestamp.
type Expirer struct {
	client    *Client
	logger    *audit.Logger
	metaKey   string
	dryRun    bool
}

// NewExpirer creates an Expirer. metaKey is the secret data key that holds an
// RFC3339 expiry timestamp (e.g. "expires_at").
func NewExpirer(client *Client, logger *audit.Logger, metaKey string, dryRun bool) *Expirer {
	return &Expirer{
		client:  client,
		logger:  logger,
		metaKey: metaKey,
		dryRun:  dryRun,
	}
}

// CheckAndPurge iterates over paths, reads each secret, and purges those whose
// expiry timestamp is in the past. It returns one ExpiryResult per path.
func (e *Expirer) CheckAndPurge(paths []string) []ExpiryResult {
	results := make([]ExpiryResult, 0, len(paths))
	for _, p := range paths {
		result := e.process(p)
		results = append(results, result)
	}
	return results
}

// Summary returns counts of expired, active, and errored secrets from a slice
// of ExpiryResults.
func Summary(results []ExpiryResult) (expired, active, errored int) {
	for _, r := range results {
		switch {
		case r.Error != nil:
			errored++
		case r.Expired:
			expired++
		default:
			active++
		}
	}
	return
}

func (e *Expirer) process(path string) ExpiryResult {
	secret, err := e.client.ReadSecret(path)
	if err != nil {
		return ExpiryResult{Path: path, Error: fmt.Errorf("read: %w", err)}
	}

	raw, ok := secret[e.metaKey]
	if !ok {
		return ExpiryResult{Path: path, Expired: false}
	}

	expStr, ok := raw.(string)
	if !ok {
		return ExpiryResult{Path: path, Error: fmt.Errorf("key %q is not a string", e.metaKey)}
	}

	expTime, err := time.Parse(time.RFC3339, expStr)
	if err != nil {
		return ExpiryResult{Path: path, Error: fmt.Errorf("parse %q: %w", expStr, err)}
	}

	if time.Now().Before(expTime) {
		return ExpiryResult{Path: path, Expired: false}
	}

	// Secret is expired.
	if e.dryRun {
		e.logger.Log("expire_dry_run", path, nil)
		return ExpiryResult{Path: path, Expired: true}
	}

	if err := e.client.DeleteSecret(path); err != nil {
		e.logger.Log("expire_delete_error", path, err)
		return ExpiryResult{Path: path, Expired: true, Error: fmt.Errorf("delete: %w", err)}
	}

	e.logger.Log("expire_deleted", path, nil)
	return ExpiryResult{Path: path, Expired: true}
}
