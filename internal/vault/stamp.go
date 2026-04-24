package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

// Stamper writes a timestamp field into secrets at a given path prefix.
type Stamper struct {
	client  *api.Client
	logger  auditLogger
	field   string
	dryRun  bool
}

type StamperOption func(*Stamper)

func WithStampDryRun(dryRun bool) StamperOption {
	return func(s *Stamper) { s.dryRun = dryRun }
}

func WithStampField(field string) StamperOption {
	return func(s *Stamper) { s.field = field }
}

// NewStamper creates a Stamper that injects a timestamp field into secrets.
func NewStamper(client *api.Client, logger auditLogger, opts ...StamperOption) (*Stamper, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	s := &Stamper{
		client: client,
		logger: logger,
		field:  "_stamped_at",
	}
	for _, o := range opts {
		o(s)
	}
	if s.field == "" {
		return nil, fmt.Errorf("stamp field name must not be empty")
	}
	return s, nil
}

// Stamp reads the secret at path, injects the timestamp field, and writes it back.
// If dry-run is enabled, the write is skipped.
func (s *Stamper) Stamp(path string) error {
	secret, err := s.client.Logical().Read(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return fmt.Errorf("secret not found: %s", path)
	}

	data := make(map[string]interface{}, len(secret.Data)+1)
	for k, v := range secret.Data {
		data[k] = v
	}
	data[s.field] = time.Now().UTC().Format(time.RFC3339)

	s.logger.Log(map[string]interface{}{
		"op":      "stamp",
		"path":    path,
		"field":   s.field,
		"dry_run": s.dryRun,
	})

	if s.dryRun {
		return nil
	}

	_, err = s.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
