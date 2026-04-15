package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformPath_StripAndAdd(t *testing.T) {
	tr := NewTransformer(TransformRule{
		StripPrefix: "old/ns/",
		AddPrefix:   "new/ns/",
	})

	out, err := tr.TransformPath("old/ns/secret/db")
	require.NoError(t, err)
	assert.Equal(t, "new/ns/secret/db", out)
}

func TestTransformPath_StripOnly(t *testing.T) {
	tr := NewTransformer(TransformRule{
		StripPrefix: "prod/",
	})

	out, err := tr.TransformPath("prod/api/key")
	require.NoError(t, err)
	assert.Equal(t, "api/key", out)
}

func TestTransformPath_AddOnly(t *testing.T) {
	tr := NewTransformer(TransformRule{
		AddPrefix: "staging/",
	})

	out, err := tr.TransformPath("api/key")
	require.NoError(t, err)
	assert.Equal(t, "staging/api/key", out)
}

func TestTransformPath_MissingPrefix_ReturnsError(t *testing.T) {
	tr := NewTransformer(TransformRule{
		StripPrefix: "prod/",
	})

	_, err := tr.TransformPath("staging/api/key")
	assert.ErrorContains(t, err, "does not have expected prefix")
}

func TestTransformPath_NoRules_Passthrough(t *testing.T) {
	tr := NewTransformer(TransformRule{})

	out, err := tr.TransformPath("any/path")
	require.NoError(t, err)
	assert.Equal(t, "any/path", out)
}

func TestTransformData_RenamesKeys(t *testing.T) {
	tr := NewTransformer(TransformRule{
		RenameKeys: map[string]string{
			"old_key": "new_key",
		},
	})

	input := map[string]interface{}{
		"old_key": "value1",
		"other":   "value2",
	}

	out := tr.TransformData(input)
	assert.Equal(t, "value1", out["new_key"])
	assert.Equal(t, "value2", out["other"])
	_, exists := out["old_key"]
	assert.False(t, exists)
}

func TestTransformData_NoRules_ReturnsSameMap(t *testing.T) {
	tr := NewTransformer(TransformRule{})

	input := map[string]interface{}{"key": "val"}
	out := tr.TransformData(input)
	assert.Equal(t, input, out)
}
