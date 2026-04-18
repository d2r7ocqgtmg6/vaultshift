package vault

import (
	"fmt"

	"github.com/vaultshift/internal/audit"
)

// Annotator adds or updates metadata annotations on secrets.
type Annotator struct {
	client  *Client
	logger  *audit.Logger
	dryRun  bool
	annotations map[string]string
}

// NewAnnotator creates an Annotator. annotations is a map of key->value pairs
// to merge into each secret's data under an "_annotations" sub-key.
func NewAnnotator(client *Client, logger *audit.Logger, annotations map[string]string, dryRun bool) (*Annotator, error) {
	if client == nil {
		return nil, fmt.Errorf("annotator: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("annotator: logger is required")
	}
	if len(annotations) == 0 {
		return nil, fmt.Errorf("annotator: at least one annotation is required")
	}
	return &Annotator{client: client, logger: logger, dryRun: dryRun, annotations: annotations}, nil
}

// Annotate reads the secret at path, merges annotations, and writes it back.
func (a *Annotator) Annotate(path string) error {
	secret, err := a.client.ReadSecret(path)
	if err != nil {
		return fmt.Errorf("annotate: read %s: %w", path, err)
	}

	data := secret
	if data == nil {
		data = make(map[string]interface{})
	}

	existing, _ := data["_annotations"].(map[string]interface{})
	if existing == nil {
		existing = make(map[string]interface{})
	}
	for k, v := range a.annotations {
		existing[k] = v
	}
	data["_annotations"] = existing

	a.logger.Log("annotate", map[string]interface{}{"path": path, "dry_run": a.dryRun})

	if a.dryRun {
		return nil
	}
	return a.client.WriteSecret(path, data)
}
