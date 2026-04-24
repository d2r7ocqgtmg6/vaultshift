package vault

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultshift/internal/audit"
)

// LintRule is a function that checks a secret path and its data, returning a
// list of violation messages (empty means no violations).
type LintRule func(path string, data map[string]interface{}) []string

// LintResult holds the outcome of linting a single secret path.
type LintResult struct {
	Path       string
	Violations []string
}

// Linter checks secrets against a set of configurable rules.
type Linter struct {
	client *Client
	logger *audit.Logger
	rules  []LintRule
}

// NewLinter creates a new Linter. At least one rule is required.
func NewLinter(client *Client, logger *audit.Logger, rules ...LintRule) (*Linter, error) {
	if client == nil {
		return nil, fmt.Errorf("lint: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("lint: logger is required")
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("lint: at least one rule is required")
	}
	return &Linter{client: client, logger: logger, rules: rules}, nil
}

// Lint reads the secret at path and evaluates all rules against it.
// It returns a LintResult and logs each violation.
func (l *Linter) Lint(path string) (LintResult, error) {
	result := LintResult{Path: path}

	data, err := l.client.ReadSecret(path)
	if err != nil {
		return result, fmt.Errorf("lint: read %q: %w", path, err)
	}

	for _, rule := range l.rules {
		violations := rule(path, data)
		result.Violations = append(result.Violations, violations...)
	}

	for _, v := range result.Violations {
		l.logger.Log("lint", map[string]interface{}{
			"path":      path,
			"violation": v,
		})
	}

	return result, nil
}

// NoEmptyKeys is a built-in LintRule that flags any key with an empty string value.
func NoEmptyKeys(path string, data map[string]interface{}) []string {
	var violations []string
	for k, v := range data {
		if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
			violations = append(violations, fmt.Sprintf("key %q has empty value", k))
		}
	}
	return violations
}

// NoUpperCaseKeys is a built-in LintRule that flags keys containing uppercase letters.
func NoUpperCaseKeys(path string, data map[string]interface{}) []string {
	var violations []string
	for k := range data {
		if k != strings.ToLower(k) {
			violations = append(violations, fmt.Sprintf("key %q contains uppercase letters", k))
		}
	}
	return violations
}
