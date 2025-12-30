package update

import (
	"archive/tar"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// Verifier handles update signature verification
type Verifier struct {
	config        *config.Config
	logger        *utils.Logger
	publicKey     ed25519.PublicKey
	publicKeyLoader *PublicKeyLoader
}

// NewVerifier creates a new verifier
func NewVerifier(cfg *config.Config, logger *utils.Logger) (*Verifier, error) {
	publicKeyLoader := NewPublicKeyLoader(cfg, logger)
	
	// Try to load public key (may fail initially, will be loaded on first use)
	publicKey, err := publicKeyLoader.LoadPublicKey()
	if err != nil {
		logger.Warn("Failed to load public key initially, will retry on verification", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue without public key - will be loaded on first verification
	}

	return &Verifier{
		config:          cfg,
		logger:          logger,
		publicKey:       publicKey,
		publicKeyLoader: publicKeyLoader,
	}, nil
}

// VerifyUpdate verifies the update package integrity and signature
func (v *Verifier) VerifyUpdate(updatePath string, manifest *UpdateManifest) error {
	v.logger.Info("Verifying update", map[string]interface{}{
		"path": updatePath,
	})

	// Ensure public key is loaded
	if v.publicKey == nil {
		var err error
		v.publicKey, err = v.publicKeyLoader.LoadPublicKey()
		if err != nil {
			return fmt.Errorf("failed to load public key: %v", err)
		}
	}

	// Extract update package
	extractDir := filepath.Join(v.config.UpdateDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %v", err)
	}

	// Extract tar.gz
	if err := extractTarGz(updatePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract update: %v", err)
	}

	// Verify each file in manifest
	for filename, expectedChecksum := range manifest.Checksums {
		filePath := filepath.Join(extractDir, filename)
		
		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file missing in update: %s", filename)
		}

		// Calculate checksum
		actualChecksum, err := calculateSHA256(filePath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %v", filename, err)
		}

		// Verify checksum
		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", 
				filename, expectedChecksum, actualChecksum)
		}

		// Verify signature if present
		if sig, ok := manifest.Signatures[filename]; ok {
			if err := v.verifySignature(filePath, sig); err != nil {
				return fmt.Errorf("signature verification failed for %s: %v", filename, err)
			}
		}
	}

	v.logger.Info("Update verification successful")
	return nil
}

// verifySignature verifies Ed25519 signature
func (v *Verifier) verifySignature(filePath, signatureHex string) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Decode signature
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %v", err)
	}

	// Verify signature
	if !ed25519.Verify(v.publicKey, data, signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// calculateSHA256 calculates SHA256 checksum of a file
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}


// extractTarGz extracts a tar.gz file
func extractTarGz(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open archive: %v", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to extract file: %v", err)
			}
			outFile.Close()
		}
	}

	return nil
}

