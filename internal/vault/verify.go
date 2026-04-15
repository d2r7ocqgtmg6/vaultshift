package vault

import (
	"fmt"
	"sort"
)

// VerifyResult holds the outcome of a post-migration verification.
type VerifyResult struct {
	Matched  []string
	Missing  []string
	Mismatch []string
}

// Verifier checks that secrets were successfully migrated from src to dst.
type Verifier struct {
	src *Client
	dst *Client
}

// NewVerifier creates a Verifier using the provided source and destination clients.
func NewVerifier(src, dst *Client) *Verifier {
	return &Verifier{src: src, dst: dst}
}

// Verify compares secrets at the given paths between source and destination.
// It returns a VerifyResult summarising matched, missing, and mismatched keys.
func (v *Verifier) Verify(paths []string) (*VerifyResult, error) {
	result := &VerifyResult{}

	for _, path := range paths {
		srcData, err := v.src.ReadSecret(path)
		if err != nil {
			return nil, fmt.Errorf("verify: read source %q: %w", path, err)
		}

		dstData, err := v.dst.ReadSecret(path)
		if err != nil {
			result.Missing = append(result.Missing, path)
			continue
		}

		if secretsEqual(srcData, dstData) {
			result.Matched = append(result.Matched, path)
		} else {
			result.Mismatch = append(result.Mismatch, path)
		}
	}

	sort.Strings(result.Matched)
	sort.Strings(result.Missing)
	sort.Strings(result.Mismatch)

	return result, nil
}

// secretsEqual performs a deep-equal comparison on two secret data maps.
func secretsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", va) != fmt.Sprintf("%v", vb) {
			return false
		}
	}
	return true
}
