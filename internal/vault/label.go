package vault

import (
	"errors"
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Labeler attaches or removes key-value labels stored under a reserved
// "_labels" key in each secret at the given paths.
type Labeler struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
	labels  map[string]string
	remove  []string
}

func NewLabeler(client *Client, logger *audit.Logger, labels map[string]string, remove []string, dryRun bool) (*Labeler, error) {
	if client == nil {
		return nil, errors.New("labeler: client is required")
	}
	if logger == nil {
		return nil, errors.New("labeler: logger is required")
	}
	if len(labels) == 0 && len(remove) == 0 {
		return nil, errors.New("labeler: at least one label or removal key is required")
	}
	return &Labeler{client: client, logger: logger, labels: labels, remove: remove, dryRun: dryRun}, nil
}

// Label applies label mutations to each provided path.
func (l *Labeler) Label(paths []string) error {
	for _, p := range paths {
		if err := l.applyLabels(p); err != nil {
			return fmt.Errorf("labeler: %s: %w", p, err)
		}
	}
	return nil
}

func (l *Labeler) applyLabels(path string) error {
	data, err := l.client.ReadSecret(path)
	if err != nil {
		return err
	}

	raw, _ := data["_labels"].(map[string]interface{})
	if raw == nil {
		raw = map[string]interface{}{}
	}

	for _, k := range l.remove {
		delete(raw, k)
	}
	for k, v := range l.labels {
		raw[k] = v
	}

	data["_labels"] = raw

	if l.dryRun {
		l.logger.Log("label", map[string]interface{}{"path": path, "dry_run": true})
		return nil
	}

	if err := l.client.WriteSecret(path, data); err != nil {
		return err
	}
	l.logger.Log("label", map[string]interface{}{"path": path, "labels": raw})
	return nil
}
