package vault

import "strings"

// Filter holds include/exclude rules for secret path filtering.
type Filter struct {
	IncludePrefixes []string
	ExcludePrefixes []string
}

// NewFilter creates a Filter with the given include and exclude prefix lists.
func NewFilter(include, exclude []string) *Filter {
	return &Filter{
		IncludePrefixes: include,
		ExcludePrefixes: exclude,
	}
}

// Allow returns true if the given path should be processed.
// Exclude rules take precedence over include rules.
// If no include prefixes are defined, all paths are included by default.
func (f *Filter) Allow(path string) bool {
	for _, ex := range f.ExcludePrefixes {
		if strings.HasPrefix(path, ex) {
			return false
		}
	}
	if len(f.IncludePrefixes) == 0 {
		return true
	}
	for _, in := range f.IncludePrefixes {
		if strings.HasPrefix(path, in) {
			return true
		}
	}
	return false
}

// FilterPaths returns only the paths that pass the filter.
func (f *Filter) FilterPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if f.Allow(p) {
			out = append(out, p)
		}
	}
	return out
}
