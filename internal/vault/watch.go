package vault

import (
	"context"
	"time"

	"github.com/yourusername/vaultshift/internal/audit"
)

// WatchEvent represents a change detected in a Vault secret path.
type WatchEvent struct {
	Path    string
	Changed bool
	Error   error
}

// Watcher polls a set of secret paths at a given interval and emits events
// when the data at a path changes compared to a previously captured snapshot.
type Watcher struct {
	client   *Client
	logger   *audit.Logger
	interval time.Duration
	paths    []string
	last     map[string]map[string]interface{}
}

// NewWatcher creates a Watcher for the given paths.
func NewWatcher(client *Client, logger *audit.Logger, interval time.Duration, paths []string) (*Watcher, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is required")
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one path is required")
	}
	return &Watcher{
		client:   client,
		logger:   logger,
		interval: interval,
		paths:    paths,
		last:     make(map[string]map[string]interface{}),
	}, nil
}

// Watch starts polling and sends WatchEvents on the returned channel until ctx
// is cancelled.
func (w *Watcher) Watch(ctx context.Context) <-chan WatchEvent {
	ch := make(chan WatchEvent, len(w.paths))
	go func() {
		defer close(ch)
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.poll(ctx, ch)
			}
		}
	}()
	return ch
 := range w.paths {
, err := w.client.ReadSecret(ctx, path)
		if err != nil {
			if w.logger != nil {
				w.logger.Log("watch_error", path, err)
			}
			ch <- WatchEvent{Path: path, Error: err}
			continue
		}
		prev, seen := w.last[path]
		changed := !seen || !secretDataEqual(prev, data)
		if changed {
			w.last[path] = data
			if w.logger != nil {
				w.logger.Log("watch_change", path, nil)
			}
		}
		ch <- WatchEvent{Path: path, Changed: changed}
	}
}

func secretDataEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || fmt.Sprintf("%v", v) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}
