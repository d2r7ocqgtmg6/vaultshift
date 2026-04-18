package vault

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// HistoryEntry represents a single version record of a secret.
type HistoryEntry struct {
	Path      string
	Version   int
	CreatedAt time.Time
	Deleted   bool
	Data      map[string]interface{}
}

// Historian reads secret version history from a KV v2 mount.
type Historian struct {
	client *Client
	mount  string
}

// NewHistorian creates a Historian for the given client and KV v2 mount.
func NewHistorian(client *Client, mount string) (*Historian, error) {
	if client == nil {
		return nil, fmt.Errorf("historian: client is required")
	}
	if mount == "" {
		return nil, fmt.Errorf("historian: mount is required")
	}
	return &Historian{client: client, mount: mount}, nil
}

// ListVersions returns all version metadata for the secret at path,
// sorted ascending by version number.
func (h *Historian) ListVersions(ctx context.Context, path string) ([]HistoryEntry, error) {
	metaPath := fmt.Sprintf("%s/metadata/%s", h.mount, path)
	secret, err := h.client.vault.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		return nil, fmt.Errorf("historian: read metadata %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("historian: no metadata found for %q", path)
	}

	versions, ok := secret.Data["versions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("historian: unexpected metadata format for %q", path)
	}

	var entries []HistoryEntry
	for vStr, vRaw := range versions {
		vMap, ok := vRaw.(map[string]interface{})
		if !ok {
			continue
		}
		var vNum int
		fmt.Sscanf(vStr, "%d", &vNum)

		entry := HistoryEntry{Path: path, Version: vNum}
		if ts, ok := vMap["created_time"].(string); ok {
			entry.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
		}
		if del, ok := vMap["deletion_time"].(string); ok && del != "" {
			entry.Deleted = true
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Version < entries[j].Version
	})

	return entries, nil
}

// ReadVersion fetches the data of a specific version of a secret.
func (h *Historian) ReadVersion(ctx context.Context, path string, version int) (*HistoryEntry, error) {
	dataPath := fmt.Sprintf("%s/data/%s", h.mount, path)
	secret, err := h.client.vault.Logical().ReadWithDataWithContext(ctx, dataPath, map[string][]string{
		"version": {fmt.Sprintf("%d", version)},
	})
	if err != nil {
		return nil, fmt.Errorf("historian: read version %d of %q: %w", version, path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("historian: version %d of %q not found", version, path)
	}

	entry := &HistoryEntry{Path: path, Version: version}
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		entry.Data = data
	}
	return entry, nil
}
