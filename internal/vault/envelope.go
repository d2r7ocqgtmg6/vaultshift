package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// EnvelopeEntry represents a single wrapped secret with metadata.
type EnvelopeEntry struct {
	Path      string            `json:"path"`
	Data      map[string]string `json:"data"`
	CreatedAt time.Time         `json:"created_at"`
	Version   int               `json:"version"`
}

// Envelope holds a collection of wrapped secret entries.
type Envelope struct {
	Entries   []EnvelopeEntry `json:"entries"`
	ExportedAt time.Time      `json:"exported_at"`
}

// Enveloper wraps secrets read from Vault into a portable envelope file.
type Enveloper struct {
	client *Client
	logger *AuditLogger
	dryRun bool
}

// NewEnveloper creates a new Enveloper. Returns an error if client or logger is nil.
func NewEnveloper(client *Client, logger *AuditLogger, dryRun bool) (*Enveloper, error) {
	if client == nil {
		return nil, fmt.Errorf("enveloper: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("enveloper: logger is required")
	}
	return &Enveloper{client: client, logger: logger, dryRun: dryRun}, nil
}

// Wrap reads secrets at the given paths and writes them to outPath as JSON.
// In dry-run mode the file is not written.
func (e *Enveloper) Wrap(paths []string, outPath string) (*Envelope, error) {
	env := &Envelope{ExportedAt: time.Now().UTC()}

	for _, p := range paths {
		data, err := e.client.ReadSecret(p)
		if err != nil {
			e.logger.Log("envelope_wrap", p, err)
			return nil, fmt.Errorf("enveloper: read %s: %w", p, err)
		}
		entry := EnvelopeEntry{
			Path:      p,
			Data:      data,
			CreatedAt: time.Now().UTC(),
			Version:   1,
		}
		env.Entries = append(env.Entries, entry)
		e.logger.Log("envelope_wrap", p, nil)
	}

	if e.dryRun {
		return env, nil
	}

	f, err := os.Create(outPath)
	if err != nil {
		return nil, fmt.Errorf("enveloper: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(env); err != nil {
		return nil, fmt.Errorf("enveloper: encode: %w", err)
	}
	return env, nil
}

// LoadEnvelope reads an envelope file from disk.
func LoadEnvelope(path string) (*Envelope, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load envelope: %w", err)
	}
	defer f.Close()
	var env Envelope
	if err := json.NewDecoder(f).Decode(&env); err != nil {
		return nil, fmt.Errorf("load envelope: decode: %w", err)
	}
	return &env, nil
}
