package vault

import (
	"fmt"
	"sort"
	"strings"
)

// IndexEntry holds metadata about a single secret path.
type IndexEntry struct {
	Path string
	Keys []string
}

// Index is a map from path to its IndexEntry.
type Index map[string]IndexEntry

// Indexer builds a searchable index of secret paths and their keys.
type Indexer struct {
	client *Client
}

// NewIndexer creates a new Indexer. Returns an error if client is nil.
func NewIndexer(client *Client) (*Indexer, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is required")
	}
	return &Indexer{client: client}, nil
}

// Build walks all secrets under prefix and returns an Index.
func (i *Indexer) Build(prefix string) (Index, error) {
	paths, err := listAllPaths(i.client, prefix)
	if err != nil {
		return nil, fmt.Errorf("listing paths: %w", err)
	}

	idx := make(Index, len(paths))
	for _, p := range paths {
		secret, err := i.client.ReadSecret(p)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		keys := make([]string, 0, len(secret))
		for k := range secret {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		idx[p] = IndexEntry{Path: p, Keys: keys}
	}
	return idx, nil
}

// Lookup returns all paths whose keys contain the given key name.
func (idx Index) Lookup(key string) []string {
	var results []string
	for path, entry := range idx {
		for _, k := range entry.Keys {
			if strings.EqualFold(k, key) {
				results = append(results, path)
				break
			}
		}
	}
	sort.Strings(results)
	return results
}
