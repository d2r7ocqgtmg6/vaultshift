package vault

import (
	"context"
	"fmt"
	"strings"
)

// ListSecrets recursively lists all secret paths under the given prefix.
func (c *Client) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.TrimSuffix(prefix, "/")
	listPath := fmt.Sprintf("%s/metadata/%s", c.mountPath, prefix)

	secret, err := c.logical.ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("listing %q: %w", listPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys format at %q", listPath)
	}

	var paths []string
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			continue
		}
		full := prefix + "/" + key
		if strings.HasSuffix(key, "/") {
			// recurse into sub-directory
			sub, err := c.ListSecrets(ctx, strings.TrimSuffix(full, "/"))
			if err != nil {
				return nil, err
			}
			paths = append(paths, sub...)
		} else {
			paths = append(paths, full)
		}
	}
	return paths, nil
}
