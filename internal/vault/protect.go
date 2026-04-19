package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Protector marks secrets as protected by writing a metadata flag,
// preventing accidental deletion or overwrite by other vaultshift commands.
type Protector struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewProtector creates a Protector. Returns an error if client or logger is nil.
func NewProtector(client *Client, logger *audit.Logger, dryRun bool) (*Protector, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Protector{client: client, logger: logger, dryRun: dryRun}, nil
}

// Protect marks the secret at path as protected.
func (p *Protector) Protect(path string) error {
	secret, err := p.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found: %s", path)
	}

	secret["_protected"] = "true"

	p.logger.Log(map[string]any{
		"action":  "protect",
		"path":    path,
		"dry_run": p.dryRun,
	})

	if p.dryRun {
		return nil
	}
	return p.client.WriteSecret(path, secret)
}

// Unprotect removes the protection flag from the secret at path.
func (p *Protector) Unprotect(path string) error {
	secret, err := p.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found: %s", path)
	}

	delete(secret, "_protected")

	p.logger.Log(map[string]any{
		"action":  "unprotect",
		"path":    path,
		"dry_run": p.dryRun,
	})

	if p.dryRun {
		return nil
	}
	return p.client.WriteSecret(path, secret)
}

// IsProtected returns true if the secret at path has the protection flag set.
func (p *Protector) IsProtected(path string) (bool, error) {
	secret, err := p.client.ReadSecret(path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	if secret == nil {
		return false, nil
	}
	return secret["_protected"] == "true", nil
}
