package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time capture of secrets at a given prefix.
type Snapshot struct {
	CapturedAt time.Time            `json:"captured_at"`
	Prefix     string               `json:"prefix"`
	Secrets    map[string]SecretData `json:"secrets"`
}

// SecretData holds the key-value pairs for a single secret path.
type SecretData struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

// Snapshotter captures secrets from a Vault client under a given prefix.
type Snapshotter struct {
	client *Client
	lister *Lister
}

// NewSnapshotter creates a new Snapshotter using the provided client.
func NewSnapshotter(c *Client) *Snapshotter {
	return &Snapshotter{
		client: c,
		lister: NewLister(c),
	}
}

// Capture walks all secret paths under prefix and returns a Snapshot.
func (s *Snapshotter) Capture(prefix string) (*Snapshot, error) {
	paths, err := s.lister.ListSecrets(prefix)
	if err != nil {
		return nil, fmt.Errorf("snapshot list: %w", err)
	}

	snap := &Snapshot{
		CapturedAt: time.Now().UTC(),
		Prefix:     prefix,
		Secrets:    make(map[string]SecretData, len(paths)),
	}

	for _, p := range paths {
		data, err := s.client.ReadSecret(p)
		if err != nil {
			return nil, fmt.Errorf("snapshot read %q: %w", p, err)
		}
		snap.Secrets[p] = SecretData{Path: p, Data: data}
	}

	return snap, nil
}

// Save writes the snapshot as JSON to the given file path.
func (s *Snapshot) Save(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("snapshot save: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("snapshot encode: %w", err)
	}
	return nil
}

// LoadSnapshot reads a snapshot from a JSON file.
func LoadSnapshot(filePath string) (*Snapshot, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("snapshot open: %w", err)
	}
	defer f.Close()

	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("snapshot decode: %w", err)
	}
	return &snap, nil
}
