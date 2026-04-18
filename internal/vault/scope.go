package vault

import (
	"errors"
	"fmt"
	"strings"
)

// Scoper restricts vault operations to a given namespace prefix.
type Scoper struct {
	client *Client
	prefix string
}

// NewScoper creates a Scoper that enforces the given prefix.
func NewScoper(client *Client, prefix string) (*Scoper, error) {
	if client == nil {
		return nil, errors.New("scoper: client is required")
	}
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		return nil, errors.New("scoper: prefix is required")
	}
	return &Scoper{client: client, prefix: prefix}, nil
}

// Enforce returns the scoped path if it falls within the prefix, or an error.
func (s *Scoper) Enforce(path string) (string, error) {
	scoped := strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(scoped, s.prefix+"/") && scoped != s.prefix {
		return "", fmt.Errorf("scoper: path %q is outside scope %q", path, s.prefix)
	}
	return scoped, nil
}

// ScopedPaths filters a list of paths to only those within the prefix.
func (s *Scoper) ScopedPaths(paths []string) []string {
	var out []string
	for _, p := range paths {
		if _, err := s.Enforce(p); err == nil {
			out = append(out, p)
		}
	}
	return out
}

// Read reads a secret, enforcing the scope.
func (s *Scoper) Read(path string) (map[string]interface{}, error) {
	if _, err := s.Enforce(path); err != nil {
		return nil, err
	}
	return s.client.ReadSecret(path)
}
