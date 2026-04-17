package vault

import (
	"fmt"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Pinner pins secrets at a given path by writing a "pinned" metadata marker.
type Pinner struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// PinResult holds the outcome of a pin operation.
type PinResult struct {
	Path    string
	Pinned  bool
	Skipped bool
	Error   error
}

// NewPinner creates a new Pinner.
func NewPinner(client *Client, logger *audit.Logger, dryRun bool) (*Pinner, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Pinner{client: client, logger: logger, dryRun: dryRun}, nil
}

// Pin marks a secret path as pinned by writing a metadata tag.
func (p *Pinner) Pin(path string) PinResult {
	secret, err := p.client.ReadSecret(path)
	if err != nil {
		p.logger.Log("pin_error", path, err.Error())
		return PinResult{Path: path, Error: err}
	}

	if secret == nil {
		err := fmt.Errorf("secret not found at path: %s", path)
		p.logger.Log("pin_error", path, err.Error())
		return PinResult{Path: path, Error: err}
	}

	if _, ok := secret["_pinned"]; ok {
		p.logger.Log("pin_skipped", path, "already pinned")
		return PinResult{Path: path, Skipped: true}
	}

	if p.dryRun {
		p.logger.Log("pin_dry_run", path, "would pin")
		return PinResult{Path: path, Pinned: false, Skipped: false}
	}

	secret["_pinned"] = "true"
	if err := p.client.WriteSecret(path, secret); err != nil {
		p.logger.Log("pin_error", path, err.Error())
		return PinResult{Path: path, Error: err}
	}

	p.logger.Log("pin_success", path, "pinned")
	return PinResult{Path: path, Pinned: true}
}
