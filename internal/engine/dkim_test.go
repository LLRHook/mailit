package engine

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDKIMKeyPair(t *testing.T) {
	t.Run("valid 2048-bit key", func(t *testing.T) {
		privPEM, pubBase64, err := GenerateDKIMKeyPair(2048)
		require.NoError(t, err)

		// Verify private key is valid PEM.
		block, _ := pem.Decode([]byte(privPEM))
		require.NotNil(t, block, "should decode PEM block")
		assert.Equal(t, "RSA PRIVATE KEY", block.Type)

		// Verify private key is parseable.
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		require.NoError(t, err)
		assert.Equal(t, 2048, privKey.N.BitLen(), "key should be 2048 bits")

		// Verify public key is valid base64 DER.
		pubDER, err := base64.StdEncoding.DecodeString(pubBase64)
		require.NoError(t, err)
		pubKeyIface, err := x509.ParsePKIXPublicKey(pubDER)
		require.NoError(t, err)

		pubKey, ok := pubKeyIface.(*rsa.PublicKey)
		require.True(t, ok, "public key should be RSA")
		assert.Equal(t, 2048, pubKey.N.BitLen())

		// Verify public key matches private key.
		assert.Equal(t, privKey.PublicKey.N, pubKey.N, "public keys should match")
		assert.Equal(t, privKey.PublicKey.E, pubKey.E, "public key exponents should match")
	})

	t.Run("valid 1024-bit key (minimum)", func(t *testing.T) {
		privPEM, pubBase64, err := GenerateDKIMKeyPair(1024)
		require.NoError(t, err)
		assert.NotEmpty(t, privPEM)
		assert.NotEmpty(t, pubBase64)

		block, _ := pem.Decode([]byte(privPEM))
		require.NotNil(t, block)
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		require.NoError(t, err)
		assert.Equal(t, 1024, privKey.N.BitLen())
	})

	t.Run("reject key size < 1024", func(t *testing.T) {
		_, _, err := GenerateDKIMKeyPair(512)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 1024 bits")
	})

	t.Run("reject zero key size", func(t *testing.T) {
		_, _, err := GenerateDKIMKeyPair(0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 1024 bits")
	})

	t.Run("reject negative key size", func(t *testing.T) {
		_, _, err := GenerateDKIMKeyPair(-1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 1024 bits")
	})
}

func TestEncryptDecryptPrivateKey(t *testing.T) {
	masterKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes

	t.Run("roundtrip encryption/decryption", func(t *testing.T) {
		original := "-----BEGIN RSA PRIVATE KEY-----\nfake-key-data\n-----END RSA PRIVATE KEY-----"

		encrypted, err := EncryptPrivateKey(original, masterKey)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, original, encrypted)

		// Encrypted output should be valid base64.
		_, err = base64.StdEncoding.DecodeString(encrypted)
		require.NoError(t, err, "encrypted output should be valid base64")

		decrypted, err := DecryptPrivateKey(encrypted, masterKey)
		require.NoError(t, err)
		assert.Equal(t, original, decrypted)
	})

	t.Run("roundtrip with real generated key", func(t *testing.T) {
		privPEM, _, err := GenerateDKIMKeyPair(1024)
		require.NoError(t, err)

		encrypted, err := EncryptPrivateKey(privPEM, masterKey)
		require.NoError(t, err)

		decrypted, err := DecryptPrivateKey(encrypted, masterKey)
		require.NoError(t, err)
		assert.Equal(t, privPEM, decrypted)
	})

	t.Run("wrong key fails decryption", func(t *testing.T) {
		original := "secret data"
		encrypted, err := EncryptPrivateKey(original, masterKey)
		require.NoError(t, err)

		wrongKey := []byte("abcdefghijklmnopqrstuvwxyz123456") // different 32-byte key
		_, err = DecryptPrivateKey(encrypted, wrongKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decrypting")
	})

	t.Run("encrypt rejects wrong key size", func(t *testing.T) {
		_, err := EncryptPrivateKey("data", []byte("too-short"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "32 bytes")
	})

	t.Run("decrypt rejects wrong key size", func(t *testing.T) {
		_, err := DecryptPrivateKey("dGVzdA==", []byte("too-short"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "32 bytes")
	})

	t.Run("decrypt rejects invalid base64", func(t *testing.T) {
		_, err := DecryptPrivateKey("not-valid-base64!!!", masterKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("decrypt rejects too-short ciphertext", func(t *testing.T) {
		// Encode a very short value as base64
		shortData := base64.StdEncoding.EncodeToString([]byte("ab"))
		_, err := DecryptPrivateKey(shortData, masterKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too short")
	})

	t.Run("each encryption produces different output (random nonce)", func(t *testing.T) {
		original := "same plaintext"
		enc1, err := EncryptPrivateKey(original, masterKey)
		require.NoError(t, err)
		enc2, err := EncryptPrivateKey(original, masterKey)
		require.NoError(t, err)
		assert.NotEqual(t, enc1, enc2, "two encryptions of same plaintext should differ due to random nonce")
	})
}

func TestParsePrivateKey(t *testing.T) {
	t.Run("valid PEM key", func(t *testing.T) {
		privPEM, _, err := GenerateDKIMKeyPair(1024)
		require.NoError(t, err)

		key, err := ParsePrivateKey(privPEM)
		require.NoError(t, err)
		require.NotNil(t, key)
		assert.IsType(t, &rsa.PrivateKey{}, key)
		assert.Equal(t, 1024, key.N.BitLen())
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		_, err := ParsePrivateKey("not a PEM block")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode PEM")
	})

	t.Run("PEM with wrong content", func(t *testing.T) {
		badPEM := "-----BEGIN RSA PRIVATE KEY-----\naW52YWxpZC1rZXktZGF0YQ==\n-----END RSA PRIVATE KEY-----"
		_, err := ParsePrivateKey(badPEM)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing private key")
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ParsePrivateKey("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode PEM")
	})
}

func TestSignMessage(t *testing.T) {
	// Generate a real key pair for signing.
	privPEM, _, err := GenerateDKIMKeyPair(2048)
	require.NoError(t, err)

	t.Run("signed message contains DKIM-Signature header", func(t *testing.T) {
		msg := &OutgoingMessage{
			From:     "sender@example.com",
			To:       []string{"recipient@example.com"},
			Subject:  "Test DKIM",
			TextBody: "This message should be DKIM signed.",
		}

		rawMessage, err := BuildMessage(msg)
		require.NoError(t, err)

		signed, err := SignMessage(rawMessage, "example.com", "mailit", privPEM)
		require.NoError(t, err)
		require.NotNil(t, signed)

		signedStr := string(signed)
		assert.Contains(t, signedStr, "DKIM-Signature:")
		assert.Contains(t, signedStr, "d=example.com")
		assert.Contains(t, signedStr, "s=mailit")

		// Original content should still be present.
		assert.Contains(t, signedStr, "From: sender@example.com")
		assert.Contains(t, signedStr, "Subject: Test DKIM")
	})

	t.Run("signed message is larger than original", func(t *testing.T) {
		msg := &OutgoingMessage{
			From:     "sender@example.com",
			To:       []string{"recipient@example.com"},
			Subject:  "Size Test",
			TextBody: "Body.",
		}

		rawMessage, err := BuildMessage(msg)
		require.NoError(t, err)

		signed, err := SignMessage(rawMessage, "example.com", "default", privPEM)
		require.NoError(t, err)
		assert.Greater(t, len(signed), len(rawMessage), "signed message should be larger")
	})

	t.Run("invalid private key PEM", func(t *testing.T) {
		rawMessage := []byte("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
		_, err := SignMessage(rawMessage, "example.com", "mailit", "invalid-pem")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing private key")
	})

	t.Run("DKIM-Signature starts the signed message", func(t *testing.T) {
		msg := &OutgoingMessage{
			From:     "test@example.com",
			To:       []string{"to@example.com"},
			Subject:  "Test",
			TextBody: "Body text",
		}

		rawMessage, err := BuildMessage(msg)
		require.NoError(t, err)

		signed, err := SignMessage(rawMessage, "example.com", "selector1", privPEM)
		require.NoError(t, err)

		// DKIM-Signature should be prepended to the beginning of the message.
		assert.True(t, strings.HasPrefix(string(signed), "DKIM-Signature:"),
			"DKIM-Signature header should be at the start of the signed message")
	})
}
