package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/yourusername/vaultshift/internal/audit"
)

// Encrypter encrypts secret values at a given path using AES-GCM before writing back.
type Encrypter struct {
	client *Client
	logger *audit.Logger
	key    []byte // must be 16, 24, or 32 bytes
	dryRun bool
}

// NewEncrypter constructs an Encrypter. key must be 16, 24, or 32 bytes for AES-128/192/256.
func NewEncrypter(client *Client, logger *audit.Logger, key []byte, dryRun bool) (*Encrypter, error) {
	if client == nil {
		return nil, errors.New("encrypt: client is required")
	}
	if logger == nil {
		return nil, errors.New("encrypt: logger is required")
	}
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("encrypt: key must be 16, 24, or 32 bytes; got %d", len(key))
	}
	return &Encrypter{client: client, logger: logger, key: key, dryRun: dryRun}, nil
}

// Encrypt reads the secret at path, encrypts each value with AES-GCM, and writes it back.
func (e *Encrypter) Encrypt(path string) error {
	secret, err := e.client.ReadSecret(path)
	if err != nil {
		e.logger.Log("encrypt_error", map[string]any{"path": path, "error": err.Error()})
		return fmt.Errorf("encrypt: read %q: %w", path, err)
	}

	encrypted := make(map[string]any, len(secret))
	for k, v := range secret {
		plain := fmt.Sprintf("%v", v)
		cipher, encErr := encryptAESGCM(e.key, []byte(plain))
		if encErr != nil {
			return fmt.Errorf("encrypt: key %q at %q: %w", k, path, encErr)
		}
		encrypted[k] = base64.StdEncoding.EncodeToString(cipher)
	}

	e.logger.Log("encrypt", map[string]any{"path": path, "dry_run": e.dryRun})

	if e.dryRun {
		return nil
	}
	return e.client.WriteSecret(path, encrypted)
}

func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
