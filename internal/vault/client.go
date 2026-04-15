package vault

import (
	"fmt"
	"net/http"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with namespace context.
type Client struct {
	api       *vaultapi.Client
	namespace string
}

// Config holds the parameters needed to create a Vault client.
type Config struct {
	Address   string
	Token     string
	Namespace string
	Timeout   time.Duration
}

// New creates a new Vault Client for the given configuration.
func New(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("vault address is required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("vault token is required")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	apiCfg := vaultapi.DefaultConfig()
	apiCfg.Address = cfg.Address
	apiCfg.HttpClient = &http.Client{Timeout: timeout}

	c, err := vaultapi.NewClient(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	c.SetToken(cfg.Token)

	if cfg.Namespace != "" {
		c.SetNamespace(cfg.Namespace)
	}

	return &Client{api: c, namespace: cfg.Namespace}, nil
}

// Namespace returns the namespace this client is scoped to.
func (c *Client) Namespace() string {
	return c.namespace
}

// ReadSecret reads a KV v2 secret at the given path.
func (c *Client) ReadSecret(path string) (map[string]interface{}, error) {
	secret, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}
	return secret.Data, nil
}

// WriteSecret writes data to the given path.
func (c *Client) WriteSecret(path string, data map[string]interface{}) error {
	_, err := c.api.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("writing secret at %q: %w", path, err)
	}
	return nil
}

// ListSecrets lists keys under the given path.
func (c *Client) ListSecrets(path string) ([]string, error) {
	secret, err := c.api.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("listing secrets at %q: %w", path, err)
	}
	if secret == nil {
		return []string{}, nil
	}
	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}
