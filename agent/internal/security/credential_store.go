package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// CredentialStore handles secure credential storage
type CredentialStore struct {
	dataDir string
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(dataDir string) *CredentialStore {
	return &CredentialStore{
		dataDir: dataDir,
	}
}

// Store stores encrypted credentials
func (cs *CredentialStore) Store(key string, value []byte) error {
	encrypted, err := cs.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %v", err)
	}

	credPath := cs.getCredentialPath(key)
	if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
		return fmt.Errorf("failed to create credential directory: %v", err)
	}

	if err := os.WriteFile(credPath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write credential: %v", err)
	}

	return nil
}

// Retrieve retrieves and decrypts credentials
func (cs *CredentialStore) Retrieve(key string) ([]byte, error) {
	credPath := cs.getCredentialPath(key)
	
	encrypted, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credential: %v", err)
	}

	decrypted, err := cs.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %v", err)
	}

	return decrypted, nil
}

// Delete removes stored credentials
func (cs *CredentialStore) Delete(key string) error {
	credPath := cs.getCredentialPath(key)
	return os.Remove(credPath)
}

// getCredentialPath returns platform-specific credential path
func (cs *CredentialStore) getCredentialPath(key string) string {
	credDir := filepath.Join(cs.dataDir, "credentials")
	return filepath.Join(credDir, key+".enc")
}

// encrypt encrypts data using AES-GCM
func (cs *CredentialStore) encrypt(plaintext []byte) ([]byte, error) {
	// Get encryption key from platform keychain or generate one
	key, err := cs.getEncryptionKey()
	if err != nil {
		return nil, err
	}

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

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// decrypt decrypts data using AES-GCM
func (cs *CredentialStore) decrypt(encrypted []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return nil, err
	}

	key, err := cs.getEncryptionKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// getEncryptionKey retrieves or generates encryption key
func (cs *CredentialStore) getEncryptionKey() ([]byte, error) {
	keyPath := filepath.Join(cs.dataDir, ".encryption_key")
	
	// Try to read existing key
	if key, err := os.ReadFile(keyPath); err == nil {
		return key, nil
	}

	// Generate new key
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	// Store key with platform-specific protection
	if err := cs.storeEncryptionKey(keyPath, key); err != nil {
		return nil, err
	}

	return key, nil
}

// storeEncryptionKey stores encryption key with platform-specific protection
func (cs *CredentialStore) storeEncryptionKey(path string, key []byte) error {
	// On Windows, could use DPAPI
	// On macOS, could use Keychain
	// On Linux, rely on file permissions (600)
	
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	perm := os.FileMode(0600)
	if runtime.GOOS == "windows" {
		perm = 0666 // Windows doesn't support Unix permissions the same way
	}

	return os.WriteFile(path, key, perm)
}

