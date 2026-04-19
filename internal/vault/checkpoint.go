package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

// Checkpoint represents a saved state of secrets at a point in time.
type Checkpoint struct {
	CreatedAt time.Time            `json:"created_at"`
	Prefix    string               `json:"prefix"`
	Secrets   map[string]api.Secret `json:"secrets"`
}

// Checkpointer saves and loads secret checkpoints to disk.
type Checkpointer struct {
	client *Client
}

// NewCheckpointer creates a new Checkpointer.
func NewCheckpointer(client *Client) (*Checkpointer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	return &Checkpointer{client: client}, nil
}

// Save reads all secrets under prefix and writes them to path as JSON.
func (c *Checkpointer) Save(prefix, path string, dryRun bool) error {
	paths, err := ListSecrets(c.client, prefix)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}

	secrets := make(map[string]api.Secret, len(paths))
	for _, p := range paths {
		s, err := c.client.ReadSecret(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		if s != nil {
			secrets[p] = *s
		}
	}

	cp := Checkpoint{
		CreatedAt: time.Now().UTC(),
		Prefix:    prefix,
		Secrets:   secrets,
	}

	if dryRun {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(cp)
}

// Load reads a checkpoint from disk.
func LoadCheckpoint(path string) (*Checkpoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var cp Checkpoint
	if err := json.NewDecoder(f).Decode(&cp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &cp, nil
}
