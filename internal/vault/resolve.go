package vault

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog"
)

// Resolver resolves secret references (e.g. "ref:secret/path#key") to their values.
type Resolver struct {
	client *Client
	logger zerolog.Logger
	dryRun bool
}

// NewResolver creates a new Resolver.
func NewResolver(client *Client, logger zerolog.Logger, dryRun bool) (*Resolver, error) {
	if client == nil {
		return nil, errors.New("resolver: client is required")
	}
	return &Resolver{client: client, logger: logger, dryRun: dryRun}, nil
}

// Resolve reads a secret at path and returns the value for the given key.
func (r *Resolver) Resolve(path, key string) (string, error) {
	if path == "" {
		return "", errors.New("resolve: path is required")
	}
	if key == "" {
		return "", errors.New("resolve: key is required")
	}

	r.logger.Debug().Str("path", path).Str("key", key).Msg("resolving secret reference")

	data, err := r.client.ReadSecret(path)
	if err != nil {
		return "", fmt.Errorf("resolve: read %q: %w", path, err)
	}

	val, ok := data[key]
	if !ok {
		return "", fmt.Errorf("resolve: key %q not found in %q", key, path)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("resolve: key %q in %q is not a string", key, path)
	}

	return str, nil
}

// ResolveMap resolves all ref values in a map, returning a new map with resolved values.
func (r *Resolver) ResolveMap(path string) (map[string]string, error) {
	data, err := r.client.ReadSecret(path)
	if err != nil {
		return nil, fmt.Errorf("resolve map: %w", err)
	}

	out := make(map[string]string, len(data))
	for k, v := range data {
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprintf("%v", v)
		}
		out[k] = s
	}
	return out, nil
}
