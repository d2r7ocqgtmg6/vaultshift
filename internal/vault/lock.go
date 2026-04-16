package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

// Locker manages advisory locks stored as secrets in Vault to prevent
// concurrent migrations on the same prefix.
type Locker struct {
	client    *api.Client
	lockPath  string
	identity  string
}

// LockMeta holds metadata written to the lock secret.
type LockMeta struct {
	Owner     string    `json:"owner"`
	AcquiredAt time.Time `json:"acquired_at"`
}

// NewLocker creates a Locker that stores a lock at lockPath in Vault.
func NewLocker(client *api.Client, lockPath, identity string) (*Locker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if lockPath == "" {
		return nil, fmt.Errorf("lockPath must not be empty")
	}
	if identity == "" {
		return nil, fmt.Errorf("identity must not be empty")
	}
	return &Locker{client: client, lockPath: lockPath, identity: identity}, nil
}

// Acquire writes a lock secret. Returns an error if a lock already exists.
func (l *Locker) Acquire() error {
	existing, err := l.client.Logical().Read(l.lockPath)
	if err != nil {
		return fmt.Errorf("checking lock at %s: %w", l.lockPath, err)
	}
	if existing != nil && existing.Data != nil {
		owner, _ := existing.Data["owner"].(string)
		return fmt.Errorf("lock already held by %q at %s", owner, l.lockPath)
	}

	_, err = l.client.Logical().Write(l.lockPath, map[string]interface{}{
		"owner":       l.identity,
		"acquired_at": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("acquiring lock at %s: %w", l.lockPath, err)
	}
	return nil
}

// Release deletes the lock secret. It is a no-op if no lock exists.
func (l *Locker) Release() error {
	_, err := l.client.Logical().Delete(l.lockPath)
	if err != nil {
		return fmt.Errorf("releasing lock at %s: %w", l.lockPath, err)
	}
	return nil
}

// IsLocked returns true if a lock secret currently exists at lockPath.
func (l *Locker) IsLocked() (bool, error) {
	secret, err := l.client.Logical().Read(l.lockPath)
	if err != nil {
		return false, fmt.Errorf("reading lock at %s: %w", l.lockPath, err)
	}
	return secret != nil && secret.Data != nil, nil
}
