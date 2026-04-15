package vault

import (
	"testing"
)

func TestNewTagger_ValidTags(t *testing.T) {
	tagger, err := NewTagger([]string{"env=prod", "team=platform"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tagger.tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tagger.tags))
	}
}

func TestNewTagger_InvalidTag(t *testing.T) {
	_, err := NewTagger([]string{"badtag"})
	if err == nil {
		t.Fatal("expected error for malformed tag, got nil")
	}
}

func TestNewTagger_EmptyKey(t *testing.T) {
	_, err := NewTagger([]string{"=value"})
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

func TestAnnotate_AddsTags(t *testing.T) {
	tagger, _ := NewTagger([]string{"env=staging"})
	data := map[string]interface{}{"username": "admin"}
	out := tagger.Annotate(data)

	if out["username"] != "admin" {
		t.Errorf("original key lost after annotation")
	}
	tags, ok := out["_tags"].(map[string]string)
	if !ok {
		t.Fatalf("expected _tags map, got %T", out["_tags"])
	}
	if tags["env"] != "staging" {
		t.Errorf("expected env=staging, got %q", tags["env"])
	}
}

func TestAnnotate_NoTags_NoTagsKey(t *testing.T) {
	tagger, _ := NewTagger([]string{})
	data := map[string]interface{}{"key": "value"}
	out := tagger.Annotate(data)
	if _, exists := out["_tags"]; exists {
		t.Error("expected no _tags key when tagger has no tags")
	}
}

func TestTagPaths_ReturnsMappedTags(t *testing.T) {
	tagger, _ := NewTagger([]string{"owner=infra"})
	paths := []string{"secret/a", "secret/b"}
	result := tagger.TagPaths(paths)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	for _, p := range paths {
		tm, ok := result[p]
		if !ok {
			t.Errorf("missing path %q in result", p)
		}
		if tm["owner"] != "infra" {
			t.Errorf("expected owner=infra for %q", p)
		}
	}
}

func TestTagPaths_EmptyTags_EmptyMaps(t *testing.T) {
	tagger, _ := NewTagger([]string{})
	result := tagger.TagPaths([]string{"secret/x"})
	if len(result["secret/x"]) != 0 {
		t.Error("expected empty tag map for path with no tags")
	}
}
