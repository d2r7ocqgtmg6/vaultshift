package vault

import (
	"fmt"
	"time"

	"github.com/vaultshift/internal/audit"
)

// TenureResult holds the result of a tenure check for a single secret path.
type TenureResult struct {
	Path      string
	CreatedAt time.Time
	AgeDays   int
	ExceedsMax bool
}

// Tenurer checks how long secrets have existed and flags those exceeding a max age.
type Tenurer struct {
	client  *Client
	logger  *audit.Logger
	maxDays int
}

// NewTenurer creates a Tenurer. maxDays is the threshold in days; secrets older
// than maxDays are flagged in the results.
func NewTenurer(client *Client, logger *audit.Logger, maxDays int) (*Tenurer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if maxDays <= 0 {
		return nil, fmt.Errorf("maxDays must be greater than zero")
	}
	return &Tenurer{client: client, logger: logger, maxDays: maxDays}, nil
}

// Check reads metadata for each path and returns TenureResult entries.
func (t *Tenurer) Check(paths []string) ([]TenureResult, error) {
	var results []TenureResult
	now := time.Now().UTC()

	for _, p := range paths {
		secret, err := t.client.ReadSecret(p)
		if err != nil {
			t.logger.Log("tenure_check_error", map[string]interface{}{
				"path":  p,
				"error": err.Error(),
			})
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}

		createdAt := extractCreatedAt(secret)
		ageDays := int(now.Sub(createdAt).Hours() / 24)
		exceeds := ageDays > t.maxDays

		t.logger.Log("tenure_checked", map[string]interface{}{
			"path":        p,
			"age_days":    ageDays,
			"exceeds_max": exceeds,
		})

		results = append(results, TenureResult{
			Path:       p,
			CreatedAt:  createdAt,
			AgeDays:    ageDays,
			ExceedsMax: exceeds,
		})
	}
	return results, nil
}

// extractCreatedAt pulls a "created_at" string from secret metadata if present.
func extractCreatedAt(secret map[string]interface{}) time.Time {
	if secret == nil {
		return time.Time{}
	}
	if raw, ok := secret["created_at"]; ok {
		if s, ok := raw.(string); ok {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Time{}
}
