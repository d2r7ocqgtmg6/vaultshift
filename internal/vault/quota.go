package vault

import (
	"fmt"

	"github.com/your-org/vaultshift/internal/audit"
)

// QuotaResult holds the result of a quota check for a single path.
type QuotaResult struct {
	Path    string
	Count   int
	Exceeds bool
}

// Quoter checks secret counts under prefixes against a maximum limit.
type Quoter struct {
	client *Client
	logger *audit.Logger
	limit  int
}

// NewQuoter creates a Quoter. Returns an error if client is nil or limit <= 0.
func NewQuoter(client *Client, logger *audit.Logger, limit int) (*Quoter, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is required")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}
	return &Quoter{client: client, logger: logger, limit: limit}, nil
}

// Check counts all secrets under each prefix and reports whether the limit is exceeded.
func (q *Quoter) Check(prefixes []string) ([]QuotaResult, error) {
	var results []QuotaResult
	for _, prefix := range prefixes {
		paths, err := listAllPaths(q.client, prefix)
		if err != nil {
			return nil, fmt.Errorf("listing %q: %w", prefix, err)
		}
		count := len(paths)
		exceeds := count > q.limit
		if q.logger != nil {
			q.logger.Log(map[string]interface{}{
				"op":      "quota_check",
				"prefix":  prefix,
				"count":   count,
				"limit":   q.limit,
				"exceeds": exceeds,
			})
		}
		results = append(results, QuotaResult{Path: prefix, Count: count, Exceeds: exceeds})
	}
	return results, nil
}
