package vault

import (
	"fmt"
	"strings"
)

// SearchResult holds a matched secret path and the matching key.
type SearchResult struct {
	Path  string `json:"path"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Searcher searches for secrets whose keys or values match a query.
type Searcher struct {
	client *Client
}

// NewSearcher creates a new Searcher.
func NewSearcher(c *Client) (*Searcher, error) {
	if c == nil {
		return nil, fmt.Errorf("client is required")
	}
	return &Searcher{client: c}, nil
}

// Search recursively lists paths under prefix and returns secrets matching query.
func (s *Searcher) Search(prefix, query string, searchValues bool) ([]SearchResult, error) {
	paths, err := listAllPaths(s.client, prefix)
	if err != nil {
		return nil, fmt.Errorf("list paths: %w", err)
	}

	var results []SearchResult
	for _, path := range paths {
		data, err := s.client.ReadSecret(path)
		if err != nil || data == nil {
			continue
		}
		for k, v := range data {
			keyMatch := strings.Contains(k, query)
			valStr := fmt.Sprintf("%v", v)
			valMatch := searchValues && strings.Contains(valStr, query)
			if keyMatch || valMatch {
				results = append(results, SearchResult{
					Path:  path,
					Key:   k,
					Value: valStr,
				})
			}
		}
	}
	return results, nil
}

// listAllPaths is a helper that recursively lists all secret paths.
func listAllPaths(c *Client, prefix string) ([]string, error) {
	keys, err := c.ListSecrets(prefix)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, k := range keys {
		full := strings.TrimRight(prefix, "/") + "/" + k
		if strings.HasSuffix(k, "/") {
			sub, err := listAllPaths(c, full)
			if err != nil {
				return nil, err
			}
			paths = append(paths, sub...)
		} else {
			paths = append(paths, full)
		}
	}
	return paths, nil
}
