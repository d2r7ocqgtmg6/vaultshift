package vault

import (
	"errors"
	"fmt"

	"github.com/subtlepseudonym/vaultshift/internal/audit"
)

// Rotator reads secrets from a source path, re-writes them with a new value
// generator function, and optionally dry-runs the operation.
type Rotator struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
	generate func(key, old string) (string, error)
}

type RotatorOption func(*Rotator)

func WithDryRun(dry bool) RotatorOption {
	return func(r *Rotator) { r.dryRun = dry }
}

func WithGenerator(fn func(key, old string) (string, error)) RotatorOption {
	return func(r *Rotator) { r.generate = fn }
}

func NewRotator(client *Client, logger *audit.Logger, opts ...RotatorOption) (*Rotator, error) {
	if client == nil {
		return nil, errors.New("rotate: client is required")
	}
	if logger == nil {
		return nil, errors.New("rotate: logger is required")
	}
	r := &Rotator{
		client: client,
		logger: logger,
		generate: func(key, old string) (string, error) {
			return old + "_rotated", nil
		},
	}
	for _, o := range opts {
		o(r)
	}
	return r, nil
}

// Rotate reads the secret at path, applies the generator to each key, and writes back.
func (r *Rotator) Rotate(path string) (map[string]string, error) {
	data, err := r.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("rotate: read %s: %w", path, err)
	}

	updated := make(map[string]string, len(data))
	for k, v := range data {
		newVal, err := r.generate(k, v)
		if err != nil {
			return nil, fmt.Errorf("rotate: generate key %s: %w", k, err)
		}
		updated[k] = newVal
	}

	if r.dryRun {
		r.logger.Log("rotate", map[string]interface{}{"path": path, "dry_run": true})
		return updated, nil
	}

	if err := r.client.WriteSecret(path, updated); err != nil {
		return nil, fmt.Errorf("rotate: write %s: %w", path, err)
	}
	r.logger.Log("rotate", map[string]interface{}{"path": path, "keys": len(updated)})
	return updated, nil
}
