package vault

import (
	"errors"
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// PolicyChecker reads a Vault policy and checks if a given path+capability is allowed.
type PolicyChecker struct {
	client *Client
	logger *audit.Logger
}

// PolicyResult holds the result of a policy check.
type PolicyResult struct {
	Path         string
	Capabilities []string
	Allowed      bool
}

// NewPolicyChecker creates a new PolicyChecker.
func NewPolicyChecker(client *Client, logger *audit.Logger) (*PolicyChecker, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	return &PolicyChecker{client: client, logger: logger}, nil
}

// Check queries the Vault sys/capabilities-self endpoint for the given paths.
func (p *PolicyChecker) Check(paths []string) ([]PolicyResult, error) {
	if len(paths) == 0 {
		return nil, errors.New("at least one path is required")
	}

	body := map[string]interface{}{"paths": paths}
	secret, err := p.client.vault.Logical().Write("sys/capabilities-self", body)
	if err != nil {
		return nil, fmt.Errorf("capabilities check failed: %w", err)
	}

	var results []PolicyResult
	for _, path := range paths {
		var caps []string
		if secret != nil && secret.Data != nil {
			if raw, ok := secret.Data[path]; ok {
				if list, ok := raw.([]interface{}); ok {
					for _, c := range list {
						if s, ok := c.(string); ok {
							caps = append(caps, s)
						}
					}
				}
			}
		}
		allowed := len(caps) > 0 && caps[0] !="
		resPath: path, Capabilities: capstp.logger.Logpolicy_check", map{
			"path":         path,
			"capabilities": caps,
			"allowed":      allowed,
		})
	}
	return results, nil
}
