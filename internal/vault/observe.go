package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/vaultshift/internal/audit"
)

// ObserveResult holds the result of a single observation for a secret path.
type ObserveResult struct {
	Path      string
	Exists    bool
	KeyCount  int
	ObservedAt time.Time
	Error     error
}

// Observer reads secrets and emits structured observation results.
type Observer struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewObserver creates an Observer. Returns an error if client or logger is nil.
func NewObserver(client *Client, logger *audit.Logger, dryRun bool) (*Observer, error) {
	if client == nil {
		return nil, fmt.Errorf("observe: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("observe: logger is required")
	}
	return &Observer{client: client, logger: logger, dryRun: dryRun}, nil
}

// Observe reads each path and returns an ObserveResult per path.
func (o *Observer) Observe(ctx context.Context, paths []string) []ObserveResult {
	results := make([]ObserveResult, 0, len(paths))
	for _, p := range paths {
		res := ObserveResult{Path: p, ObservedAt: time.Now().UTC()}
		data, err := o.client.ReadSecret(ctx, p)
		if err != nil {
			res.Error = err
			o.logger.Log("observe", p, "error", err.Error())
		} else if data == nil {
			res.Exists = false
			o.logger.Log("observe", p, "status", "not_found")
		} else {
			res.Exists = true
			res.KeyCount = len(data)
			o.logger.Log("observe", p, "status", "found")
		}
		results = append(results, res)
	}
	return results
}
