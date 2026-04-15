package vault

import (
	"fmt"
	"strings"
)

// TransformRule defines a key renaming or prefix substitution rule.
type TransformRule struct {
	StripPrefix  string
	AddPrefix    string
	RenameKeys   map[string]string
}

// Transformer applies transformation rules to secret paths and data.
type Transformer struct {
	rule TransformRule
}

// NewTransformer creates a Transformer with the given rule.
func NewTransformer(rule TransformRule) *Transformer {
	return &Transformer{rule: rule}
}

// TransformPath applies prefix stripping and addition to a secret path.
func (t *Transformer) TransformPath(path string) (string, error) {
	result := path

	if t.rule.StripPrefix != "" {
		if !strings.HasPrefix(result, t.rule.StripPrefix) {
			return "", fmt.Errorf("path %q does not have expected prefix %q", path, t.rule.StripPrefix)
		}
		result = strings.TrimPrefix(result, t.rule.StripPrefix)
	}

	if t.rule.AddPrefix != "" {
		result = t.rule.AddPrefix + result
	}

	return result, nil
}

// TransformData applies key renames to a secret's data map.
func (t *Transformer) TransformData(data map[string]interface{}) map[string]interface{} {
	if len(t.rule.RenameKeys) == 0 {
		return data
	}

	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		if newKey, ok := t.rule.RenameKeys[k]; ok {
			result[newKey] = v
		} else {
			result[k] = v
		}
	}
	return result
}
