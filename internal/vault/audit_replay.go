package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/vaultshift/internal/audit"
)

// ReplayEntry represents a single replayed audit log entry.
type ReplayEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Op        string            `json:"op"`
	Path      string            `json:"path"`
	Status    string            `json:"status"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Replayer reads a JSONL audit log and re-applies write operations.
type Replayer struct {
	client *Client
	logger *audit.Logger
	dryRun bool
}

// NewReplayer constructs a Replayer.
func NewReplayer(client *Client, logger *audit.Logger, dryRun bool) (*Replayer, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Replayer{client: client, logger: logger, dryRun: dryRun}, nil
}

// Replay reads entries from filePath and re-applies successful write ops.
func (r *Replayer) Replay(filePath string) (int, []error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, []error{fmt.Errorf("open audit file: %w", err)}
	}
	defer f.Close()

	var applied int
	var errs []error
	dec := json.NewDecoder(f)
	for dec.More() {
		var entry ReplayEntry
		if err := dec.Decode(&entry); err != nil {
			errs = append(errs, fmt.Errorf("decode entry: %w", err))
			continue
		}
		if entry.Op != "write" || entry.Status != "success" {
			continue
		}
		if r.dryRun {
			r.logger.Log("replay", entry.Path, "dry-run", nil)
			applied++
			continue
		}
		data := map[string]interface{}{}
		for k, v := range entry.Meta {
			data[k] = v
		}
		if err := r.client.WriteSecret(entry.Path, data); err != nil {
			errs = append(errs, fmt.Errorf("replay %s: %w", entry.Path, err))
			r.logger.Log("replay", entry.Path, "error", err)
			continue
		}
		r.logger.Log("replay", entry.Path, "success", nil)
		applied++
	}
	return applied, errs
}
