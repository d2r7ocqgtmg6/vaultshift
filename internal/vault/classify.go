package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// ClassifyResult holds the classification label for a secret path.
type ClassifyResult struct {
	Path  string
	Label string
}

// Classifier reads secrets and assigns labels based on key name patterns.
type Classifier struct {
	client *Client
	logger *audit.Logger
	rules  map[string]string // key substring -> label
}

// NewClassifier creates a Classifier. rules maps key substrings to labels.
func NewClassifier(client *Client, logger *audit.Logger, rules map[string]string) (*Classifier, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("at least one classification rule is required")
	}
	return &Classifier{client: client, logger: logger, rules: rules}, nil
}

// Classify reads the secret at path and returns a ClassifyResult.
func (c *Classifier) Classify(path string) (*ClassifyResult, error) {
	data, err := c.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	label := c.matchLabel(data)
	c.logger.Log("classify", map[string]any{"path": path, "label": label})
	return &ClassifyResult{Path: path, Label: label}, nil
}

// ClassifyAll classifies all paths and returns results.
func (c *Classifier) ClassifyAll(paths []string) ([]ClassifyResult, error) {
	var results []ClassifyResult
	for _, p := range paths {
		r, err := c.Classify(p)
		if err != nil {
			return results, err
		}
		results = append(results, *r)
	}
	return results, nil
}

func (c *Classifier) matchLabel(data map[string]any) string {
	for key := range data {
		lower := strings.ToLower(key)
		for substr, label := range c.rules {
			if strings.Contains(lower, strings.ToLower(substr)) {
				return label
			}
		}
	}
	return "unclassified"
}
