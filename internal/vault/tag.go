package vault

import (
	"fmt"
	"strings"
)

// Tag represents a key=value metadata label attached to a secret path.
type Tag struct {
	Key   string
	Value string
}

// Tagger applies metadata tags to secret paths, producing an annotated map.
type Tagger struct {
	tags []Tag
}

// NewTagger creates a Tagger from a slice of "key=value" strings.
// Returns an error if any entry is malformed.
func NewTagger(raw []string) (*Tagger, error) {
	tags := make([]Tag, 0, len(raw))
	for _, r := range raw {
		parts := strings.SplitN(r, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid tag %q: must be key=value", r)
		}
		tags = append(tags, Tag{Key: parts[0], Value: parts[1]})
	}
	return &Tagger{tags: tags}, nil
}

// Annotate returns a copy of the provided secret data map with tag entries
// merged in under the reserved "_tags" sub-map key.
func (t *Tagger) Annotate(data map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		out[k] = v
	}
	if len(t.tags) == 0 {
		return out
	}
	tagMap := make(map[string]string, len(t.tags))
	for _, tag := range t.tags {
		tagMap[tag.Key] = tag.Value
	}
	out["_tags"] = tagMap
	return out
}

// TagPaths returns a map of path -> tag annotations for a list of secret paths.
func (t *Tagger) TagPaths(paths []string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(paths))
	for _, p := range paths {
		if len(t.tags) == 0 {
			result[p] = map[string]string{}
			continue
		}
		tm := make(map[string]string, len(t.tags))
		for _, tag := range t.tags {
			tm[tag.Key] = tag.Value
		}
		result[p] = tm
	}
	return result
}
