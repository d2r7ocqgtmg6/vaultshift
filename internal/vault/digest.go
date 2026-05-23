package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/wearevault/vaultshift/internal/audit"
)

// DigestResult holds the computed digest for a single secret path.
type DigestResult struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// Digester computes deterministic SHA-256 digests for Vault secrets.
type Digester struct {
	client *Client
	logger *audit.Logger
}

// NewDigester creates a new Digester.
func NewDigester(client *Client, logger *audit.Logger) (*Digester, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Digester{client: client, logger: logger}, nil
}

// Digest reads the secret at path and returns its SHA-256 digest.
func (d *Digester) Digest(path string) (*DigestResult, error) {
	data, err := d.client.ReadSecret(path)
	if err != nil {
		d.logger.Log("digest", path, "error", err.Error())
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	digest, err := computeDigest(data)
	if err != nil {
		return nil, fmt.Errorf("compute digest %s: %w", path, err)
	}
	d.logger.Log("digest", path, "ok", digest)
	return &DigestResult{Path: path, Digest: digest}, nil
}

// DigestAll computes digests for all secrets under prefix.
func (d *Digester) DigestAll(prefix string) ([]DigestResult, error) {
	paths, err := listAllPaths(d.client, prefix)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", prefix, err)
	}
	var results []DigestResult
	for _, p := range paths {
		res, err := d.Digest(p)
		if err != nil {
			return nil, err
		}
		results = append(results, *res)
	}
	return results, nil
}

// computeDigest produces a deterministic SHA-256 hex digest from a secret map.
func computeDigest(data map[string]interface{}) (string, error) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]interface{}, len(data))
	for _, k := range keys {
		ordered[k] = data[k]
	}
	b, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
