package engine

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/emersion/go-msgauth/dkim"
)

// GenerateDKIMKeyPair generates a new RSA key pair for DKIM signing.
// It returns the private key in PEM format and the public key as a base64-encoded
// DER string suitable for inclusion in a DNS TXT record.
func GenerateDKIMKeyPair(bits int) (privateKeyPEM string, publicKeyBase64 string, err error) {
	if bits < 1024 {
		return "", "", fmt.Errorf("key size must be at least 1024 bits, got %d", bits)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("generating RSA key: %w", err)
	}

	// Encode private key to PEM.
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	// Encode public key as base64 DER for DNS TXT record.
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("marshaling public key: %w", err)
	}
	pubBase64 := base64.StdEncoding.EncodeToString(pubBytes)

	return string(privPEM), pubBase64, nil
}

// EncryptPrivateKey encrypts a PEM-encoded private key using AES-256-GCM.
// The master key must be exactly 32 bytes for AES-256.
func EncryptPrivateKey(plaintext string, masterKey []byte) (string, error) {
	if len(masterKey) != 32 {
		return "", fmt.Errorf("master key must be 32 bytes for AES-256, got %d", len(masterKey))
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	// Seal prepends the nonce to the ciphertext for easy extraction during decryption.
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey decrypts an AES-256-GCM encrypted private key.
// The master key must be exactly 32 bytes for AES-256.
func DecryptPrivateKey(encrypted string, masterKey []byte) (string, error) {
	if len(masterKey) != 32 {
		return "", fmt.Errorf("master key must be 32 bytes for AES-256, got %d", len(masterKey))
	}

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("decoding base64: %w", err)
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short: %d bytes, need at least %d", len(data), nonceSize)
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}

	return string(plaintext), nil
}

// ParsePrivateKey parses a PEM-encoded RSA private key.
func ParsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	return privateKey, nil
}

// SignMessage signs an email message with DKIM. It reads the raw RFC 5322
// message, signs it, and returns the complete message with the DKIM-Signature
// header prepended.
func SignMessage(message []byte, domain, selector string, privateKeyPEM string) ([]byte, error) {
	privateKey, err := ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing private key for DKIM: %w", err)
	}

	options := &dkim.SignOptions{
		Domain:   domain,
		Selector: selector,
		Signer:   privateKey,
		Hash:     crypto.SHA256,
		HeaderKeys: []string{
			"From", "To", "Subject", "Date", "Message-ID",
			"MIME-Version", "Content-Type",
		},
	}

	var signed bytes.Buffer
	if err := dkim.Sign(&signed, bytes.NewReader(message), options); err != nil {
		return nil, fmt.Errorf("signing message with DKIM: %w", err)
	}

	return signed.Bytes(), nil
}
