package vault

import (
	"fmt"
	"sync"
)

// PrefetchResult holds the result of prefetching a single secret path.
type PrefetchResult struct {
	Path string
	Data map[string]interface{}
	Err  error
}

// Prefetcher concurrently reads a list of secret paths from Vault.
type Prefetcher struct {
	client      *Client
	concurrency int
}

// NewPrefetcher creates a Prefetcher. concurrency controls how many reads run
// in parallel; it must be >= 1.
func NewPrefetcher(client *Client, concurrency int) (*Prefetcher, error) {
	if client == nil {
		return nil, fmt.Errorf("prefetcher: client is required")
	}
	if concurrency < 1 {
		return nil, fmt.Errorf("prefetcher: concurrency must be at least 1")
	}
	return &Prefetcher{client: client, concurrency: concurrency}, nil
}

// Fetch reads all paths concurrently and returns one PrefetchResult per path.
// Results are returned in the same order as the input paths.
func (p *Prefetcher) Fetch(paths []string) []PrefetchResult {
	results := make([]PrefetchResult, len(paths))

	sem := make(chan struct{}, p.concurrency)
	var wg sync.WaitGroup

	for i, path := range paths {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, secretPath string) {
			defer wg.Done()
			defer func() { <-sem }()

			data, err := p.client.ReadSecret(secretPath)
			results[idx] = PrefetchResult{
				Path: secretPath,
				Data: data,
				Err:  err,
			}
		}(i, path)
	}

	wg.Wait()
	return results
}

// FetchMap is a convenience wrapper that returns a map of path -> data for
// every successfully fetched secret. Errors are collected and returned
// separately.
func (p *Prefetcher) FetchMap(paths []string) (map[string]map[string]interface{}, []error) {
	results := p.Fetch(paths)
	out := make(map[string]map[string]interface{}, len(paths))
	var errs []error
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.Path, r.Err))
			continue
		}
		out[r.Path] = r.Data
	}
	return out, errs
}
