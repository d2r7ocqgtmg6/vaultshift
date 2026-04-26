package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/subtlepseudonym/vaultshift/internal/audit"
)

// DigestResult holds the computed digest for a single secret path.
type DigestResult struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// Digester computes deterministic SHA-256 digests over secret data at given paths.
// Digests are stable regardless of key insertion order, making them suitable for
// change detection and integrity checks.
type Digester struct {
	client *Client
	logger *audit.Logger
}

// NewDigester constructs a Digester. Both client and logger are required.
func NewDigester(client *Client, logger *audit.Logger) (*Digester, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Digester{client: client, logger: logger}, nil
}

// Digest reads the secret at path and returns a DigestResult containing a
// deterministic hex-encoded SHA-256 digest of the sorted key-value pairs.
func (d *Digester) Digest(path string) (DigestResult, error) {
	data, err := d.client.ReadSecret(path)
	if err != nil {
		d.logger.Log("digest", path, "error", err.Error())
		return DigestResult{}, fmt.Errorf("read %s: %w", path, err)
	}

	digest, err := computeDigest(data)
	if err != nil {
		d.logger.Log("digest", path, "error", err.Error())
		return DigestResult{}, fmt.Errorf("compute digest for %s: %w", path, err)
	}

	result := DigestResult{Path: path, Digest: digest}
	d.logger.Log("digest", path, "status", "ok", "digest", digest)
	return result, nil
}

// DigestAll computes digests for every path in paths and returns all results.
// Errors for individual paths are recorded in the audit log but do not abort
// processing of remaining paths.
func (d *Digester) DigestAll(paths []string) ([]DigestResult, []error) {
	results := make([]DigestResult, 0, len(paths))
	var errs []error

	for _, p := range paths {
		r, err := d.Digest(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, r)
	}

	return results, errs
}

// computeDigest builds a deterministic digest from a map of secret key-value pairs.
// Keys are sorted before hashing so that insertion order does not affect the result.
func computeDigest(data map[string]interface{}) (string, error) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		v := data[k]
		encoded, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("marshal value for key %q: %w", k, err)
		}
		// Write key=value\n into the hash for unambiguous separation.
		_, _ = fmt.Fprintf(h, "%s=%s\n", k, encoded)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
