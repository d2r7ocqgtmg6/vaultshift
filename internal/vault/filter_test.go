package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_Allow_NoRules_AllowsAll(t *testing.T) {
	f := NewFilter(nil, nil)
	assert.True(t, f.Allow("secret/foo"))
	assert.True(t, f.Allow("secret/bar/baz"))
}

func TestFilter_Allow_IncludePrefix(t *testing.T) {
	f := NewFilter([]string{"secret/prod"}, nil)
	assert.True(t, f.Allow("secret/prod/db"))
	assert.False(t, f.Allow("secret/staging/db"))
}

func TestFilter_Allow_ExcludePrefix(t *testing.T) {
	f := NewFilter(nil, []string{"secret/internal"})
	assert.True(t, f.Allow("secret/prod/db"))
	assert.False(t, f.Allow("secret/internal/keys"))
}

func TestFilter_Allow_ExcludeTakesPrecedence(t *testing.T) {
	f := NewFilter([]string{"secret/prod"}, []string{"secret/prod/internal"})
	assert.True(t, f.Allow("secret/prod/db"))
	assert.False(t, f.Allow("secret/prod/internal/key"))
}

func TestFilter_FilterPaths(t *testing.T) {
	f := NewFilter([]string{"secret/prod"}, []string{"secret/prod/skip"})
	input := []string{
		"secret/prod/db",
		"secret/prod/skip/me",
		"secret/staging/db",
		"secret/prod/api",
	}
	got := f.FilterPaths(input)
	assert.Equal(t, []string{"secret/prod/db", "secret/prod/api"}, got)
}

func TestFilter_FilterPaths_Empty(t *testing.T) {
	f := NewFilter(nil, nil)
	got := f.FilterPaths([]string{})
	assert.Empty(t, got)
}
