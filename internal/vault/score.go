package vault

import (
	"fmt"
	"strings"
)

// ScoreResult holds the computed health score for a single secret path.
type ScoreResult struct {
	Path   string
	Score  int
	Issues []string
}

// Scorer evaluates secret health based on configurable criteria.
type Scorer struct {
	client    *Client
	logger    AuditLogger
	maxAge    int // days; 0 means no age check
	required  []string
}

// NewScorer constructs a Scorer. Returns an error if client or logger is nil.
func NewScorer(client *Client, logger AuditLogger, maxAgeDays int, requiredKeys []string) (*Scorer, error) {
	if client == nil {
		return nil, fmt.Errorf("scorer: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("scorer: logger is required")
	}
	return &Scorer{
		client:   client,
		logger:   logger,
		maxAge:   maxAgeDays,
		required: requiredKeys,
	}, nil
}

// Score evaluates a single secret at path and returns a ScoreResult.
// A perfect score is 100; each issue deducts points.
func (s *Scorer) Score(path string) (*ScoreResult, error) {
	secret, err := s.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("scorer: read %s: %w", path, err)
	}

	result := &ScoreResult{Path: path, Score: 100}

	data, _ := secret["data"].(map[string]interface{})

	// Check for empty values.
	for k, v := range data {
		if str, ok := v.(string); ok && strings.TrimSpace(str) == "" {
			result.Issues = append(result.Issues, fmt.Sprintf("empty value for key %q", k))
			result.Score -= 10
		}
	}

	// Check required keys.
	for _, rk := range s.required {
		if _, ok := data[rk]; !ok {
			result.Issues = append(result.Issues, fmt.Sprintf("missing required key %q", rk))
			result.Score -= 20
		}
	}

	if result.Score < 0 {
		result.Score = 0
	}

	s.logger.Log(map[string]interface{}{
		"op":     "score",
		"path":   path,
		"score":  result.Score,
		"issues": len(result.Issues),
	})

	return result, nil
}
