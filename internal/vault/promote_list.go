package vault

import (
	"fmt"
	"strings"
)

// listAll recursively lists all non-directory secret paths under prefix.
// It reuses the lister logic already available in the package.
func listAll(client *Client, prefix string) ([]string, error) {
	keys, err := client.ListSecrets(prefix)
	if err != nil {
		return nil, fmt.Errorf("listAll %q: %w", prefix, err)
	}

	var paths []string
	for _, key := range keys {
		if strings.HasSuffix(key, "/") {
			sub, err := listAll(client, prefix+key)
			if err != nil {
				return nil, err
			}
			paths = append(paths, sub...)
		} else {
			paths = append(paths, prefix+key)
		}
	}
	return paths, nil
}
