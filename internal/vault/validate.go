package vault

import (
	"fmt"
	"strings"

	"github.com/vaultshift/internal/audit"
)

// ValidationResult holds the outcome of a single secret validation.
type ValidationResult struct {
	Path   string
	Passed bool
	Reason string
}

// Validator checks secrets against user-defined rules.
type Validator struct {
	client    *Client
	logger    *audit.Logger
	rules     []ValidationRule
	dryRun    bool
}

// ValidationRule is a named predicate applied to secret data.
type ValidationRule struct {
	Name    string
	Check   func(path string, data map[string]interface{}) (bool, string)
}

// NewValidator constructs a Validator.
func NewValidator(client *Client, logger *audit.Logger, rules []ValidationRule, dryRun bool) (*Validator, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("at least one rule is required")
	}
	return &Validator{client: client, logger: logger, rules: rules, dryRun: dryRun}, nil
}

// Validate reads the secret at path and applies all rules.
func (v *Validator) Validate(path string) ([]ValidationResult, error) {
	data, err := v.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var results []ValidationResult
	for _, rule := range v.rules {
		passed, reason := rule.Check(path, data)
		results = append(results, ValidationResult{Path: path, Passed: passed, Reason: reason})
		status := "pass"
		if !passed {
			status = "fail"
		}
		v.logger.Log(map[string]interface{}{
			"op":     "validate",
			"path":   path,
			"rule":   rule.Name,
			"status": status,
			"reason": reason,
			"dryRun": v.dryRun,
		})
	}
	return results, nil
}

// RequiredKeys returns a rule that ensures all keys are present.
func RequiredKeys(keys ...string) ValidationRule {
	return ValidationRule{
		Name: "required_keys",
		Check: func(_ string, data map[string]interface{}) (bool, string) {
			var missing []string
			for _, k := range keys {
				if _, ok := data[k]; !ok {
					missing = append(missing, k)
				}
			}
			if len(missing) > 0 {
				return false, "missing keys: " + strings.Join(missing, ", ")
			}
			return true, ""
		},
	}
}

// NoEmptyValues returns a rule that ensures no value is an empty string.
func NoEmptyValues() ValidationRule {
	return ValidationRule{
		Name: "no_empty_values",
		Check: func(_ string, data map[string]interface{}) (bool, string) {
			for k, v := range data {
				if s, ok := v.(string); ok && s == "" {
					return false, fmt.Sprintf("key %q has empty value", k)
				}
			}
			return true, ""
		},
	}
}
