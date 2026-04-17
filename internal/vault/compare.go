package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// CompareResult holds the comparison outcome for a single secret path.
type CompareResult struct {
	Path    string
	SrcData map[string]interface{}
	DstData map[string]interface{}
	Match   bool
	Missing string // "src", "dst", or ""
}

// Comparer compares secrets between two Vault clients.
type Comparer struct {
	src *api.Client
	dst *api.Client
}

// NewComparer creates a new Comparer. Returns error if either client is nil.
func NewComparer(src, dst *api.Client) (*Comparer, error) {
	if src == nil {
		return nil, fmt.Errorf("source client is required")
	}
	if dst == nil {
		return nil, fmt.Errorf("destination client is required")
	}
	return &Comparer{src: src, dst: dst}, nil
}

// Compare reads the secret at path from both src and dst and returns a CompareResult.
func (c *Comparer) Compare(path string) (*CompareResult, error) {
	srcSecret, err := c.src.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("read src %s: %w", path, err)
	}
	dstSecret, err := c.dst.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("read dst %s: %w", path, err)
	}

	res := &CompareResult{Path: path}

	if srcSecret == nil {
		res.Missing = "src"
		return res, nil
	}
	if dstSecret == nil {
		res.Missing = "dst"
		res.SrcData = srcSecret.Data
		return res, nil
	}

	res.SrcData = srcSecret.Data
	res.DstData = dstSecret.Data
	res.Match = mapsEqual(srcSecret.Data, dstSecret.Data)
	return res, nil
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok || fmt.Sprintf("%v", v) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}
