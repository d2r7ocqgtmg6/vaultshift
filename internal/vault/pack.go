package vault

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

// Packer bundles multiple secret paths into a single JSON file.
type Packer struct {
	client  *api.Client
	logger  AuditLogger
	dryRun  bool
}

type PackEntry struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

// NewPacker returns a Packer or an error if client or logger is nil.
func NewPacker(client *api.Client, logger AuditLogger, dryRun bool) (*Packer, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("audit logger is required")
	}
	return &Packer{client: client, logger: logger, dryRun: dryRun}, nil
}

// Pack reads secrets at each path and writes them to outFile as JSONL.
func (p *Packer) Pack(paths []string, outFile string) error {
	var entries []PackEntry

	for _, path := range paths {
		secret, err := p.client.Logical().Read(path)
		if err != nil {
			p.logger.Log(map[string]interface{}{"op": "pack", "path": path, "status": "error", "error": err.Error()})
			return fmt.Errorf("read %s: %w", path, err)
		}
		data := map[string]interface{}{}
		if secret != nil && secret.Data != nil {
			data = secret.Data
		}
		entries = append(entries, PackEntry{Path: path, Data: data})
		p.logger.Log(map[string]interface{}{"op": "pack", "path": path, "status": "read"})
	}

	if p.dryRun {
		p.logger.Log(map[string]interface{}{"op": "pack", "file": outFile, "status": "dry-run"})
		return nil
	}

	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("create file %s: %w", outFile, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return fmt.Errorf("encode pack: %w", err)
	}

	p.logger.Log(map[string]interface{}{"op": "pack", "file": outFile, "count": len(entries), "status": "written"})
	return nil
}
