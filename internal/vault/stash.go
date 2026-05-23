package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StashEntry holds a temporarily saved secret path and its data.
type StashEntry struct {
	Path      string                 `json:"path"`
	Data      map[string]interface{} `json:"data"`
	StashedAt time.Time              `json:"stashed_at"`
}

// Stasher temporarily saves secrets to a local file for later restoration.
type Stasher struct {
	client *Client
	logger AuditLogger
	dryRun bool
}

// NewStasher creates a Stasher. Returns an error if client or logger is nil.
func NewStasher(client *Client, logger AuditLogger, dryRun bool) (*Stasher, error) {
	if client == nil {
		return nil, fmt.Errorf("stasher: client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("stasher: logger is required")
	}
	return &Stasher{client: client, logger: logger, dryRun: dryRun}, nil
}

// Stash reads secrets at the given paths and writes them to outPath as JSONL.
// In dry-run mode no file is written.
func (s *Stasher) Stash(paths []string, outPath string) ([]StashEntry, error) {
	var entries []StashEntry
	for _, p := range paths {
		data, err := s.client.ReadSecret(p)
		if err != nil {
			s.logger.Log("stash", p, "error", err.Error())
			return nil, fmt.Errorf("stash: read %s: %w", p, err)
		}
		entries = append(entries, StashEntry{Path: p, Data: data, StashedAt: time.Now().UTC()})
		s.logger.Log("stash", p, "ok", "")
	}
	if s.dryRun {
		return entries, nil
	}
	f, err := os.Create(outPath)
	if err != nil {
		return nil, fmt.Errorf("stash: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			return nil, fmt.Errorf("stash: encode entry: %w", err)
		}
	}
	return entries, nil
}

// LoadStash reads a JSONL stash file and returns its entries.
func LoadStash(path string) ([]StashEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load stash: %w", err)
	}
	defer f.Close()
	var entries []StashEntry
	dec := json.NewDecoder(f)
	for dec.More() {
		var e StashEntry
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("load stash: decode: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
