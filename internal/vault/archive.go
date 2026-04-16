package vault

import (
	"fmt"
	"time"

	"github.com/vaultshift/internal/audit"
)

// Archiver moves secrets to an archive prefix with a timestamp suffix.
type Archiver struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewArchiver creates a new Archiver.
func NewArchiver(client *Client, logger *audit.Logger, dryRun bool) (*Archiver, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Archiver{client: client, logger: logger, dryRun: dryRun}, nil
}

// Archive reads the secret at path, writes it to archivePrefix/path@timestamp,
// then deletes the original (unless dry-run).
func (a *Archiver) Archive(path, archivePrefix string) error {
	data, err := a.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	dest := fmt.Sprintf("%s/%s@%s", archivePrefix, path, timestamp)

	a.logger.Log(map[string]any{
		"action":  "archive",
		"source":  path,
		"dest":    dest,
		"dry_run": a.dryRun,
	})

	if a.dryRun {
		return nil
	}

	if err := a.client.WriteSecret(dest, data); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}

	if err := a.client.DeleteSecret(path); err != nil {
		return fmt.Errorf("delete %s: %w", path, err)
	}

	return nil
}
