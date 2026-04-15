package vault

import "fmt"

// DiffResult holds the comparison result between source and destination secrets.
type DiffResult struct {
	Key       string
	Status    DiffStatus
	SourceVal string
	DestVal   string
}

// DiffStatus represents the state of a secret key comparison.
type DiffStatus string

const (
	DiffStatusNew      DiffStatus = "new"      // exists in source, not in dest
	DiffStatusChanged  DiffStatus = "changed"  // exists in both, values differ
	DiffStatusUnchanged DiffStatus = "unchanged" // exists in both, values match
	DiffStatusOrphaned DiffStatus = "orphaned" // exists in dest, not in source
)

// Differ compares secrets between source and destination Vault clients.
type Differ struct {
	src *Client
	dst *Client
}

// NewDiffer creates a new Differ.
func NewDiffer(src, dst *Client) *Differ {
	return &Differ{src: src, dst: dst}
}

// Diff compares a single secret key between source and destination.
func (d *Differ) Diff(path string) (*DiffResult, error) {
	srcSecret, srcErr := d.src.ReadSecret(path)
	dstSecret, dstErr := d.dst.ReadSecret(path)

	srcMissing := srcErr != nil
	dstMissing := dstErr != nil

	if srcMissing && dstMissing {
		return nil, fmt.Errorf("secret %q not found in source or destination", path)
	}

	if srcMissing {
		return &DiffResult{Key: path, Status: DiffStatusOrphaned, DestVal: fmt.Sprintf("%v", dstSecret)}, nil
	}

	if dstMissing {
		return &DiffResult{Key: path, Status: DiffStatusNew, SourceVal: fmt.Sprintf("%v", srcSecret)}, nil
	}

	srcStr := fmt.Sprintf("%v", srcSecret)
	dstStr := fmt.Sprintf("%v", dstSecret)

	if srcStr == dstStr {
		return &DiffResult{Key: path, Status: DiffStatusUnchanged, SourceVal: srcStr, DestVal: dstStr}, nil
	}

	return &DiffResult{Key: path, Status: DiffStatusChanged, SourceVal: srcStr, DestVal: dstStr}, nil
}

// DiffAll compares all secrets under a prefix between source and destination.
func (d *Differ) DiffAll(prefix string) ([]*DiffResult, error) {
	lister := &lister{client: d.src}
	keys, err := lister.ListSecrets(prefix)
	if err != nil {
		return nil, fmt.Errorf("listing source secrets: %w", err)
	}

	results := make([]*DiffResult, 0, len(keys))
	for _, key := range keys {
		result, err := d.Diff(key)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}
