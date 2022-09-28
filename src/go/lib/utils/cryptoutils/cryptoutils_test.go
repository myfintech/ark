package cryptoutils

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	commonSecret = "SUPER_SECRET"

	// openssl ran -hex 16
	commonIV = "91bdf1c65c775c8713ef1c801be6fae0"

	// openssl ran -hex 12
	commonIV12 = "8cd198ba42ca87a3266720f7"

	// openssl rand -hex 32
	commonKey = "6d4609595f397fdf16a342c86b1bef147ae7c608df95cddc42800cf912ac1f64"
)

func TestAES256CBCEncrypt(t *testing.T) {
	iv, err := hex.DecodeString(commonIV)
	require.NoError(t, err)

	encryptionKey, err := hex.DecodeString(commonKey)
	require.NoError(t, err)

	expectedHashedPassword := "UIEAF8tALbXU451yQ4TMDa6D/sS4OLeW8xaphJztNjk="
	actualHashedPassword, err := SHA256Sum(commonSecret, "base64")
	require.NoError(t, err)
	require.Equal(t, expectedHashedPassword, actualHashedPassword)

	// Generated with openssl aes-256-cbc -e -a -A -nosalt -K -iv
	expectedCipherText := "PrJfryMclBLnkjN6xz9XUQ/+uB+tm3jz1/4QTvG59zMqMTZHQYb/x0d55maaXmpP"
	cipherTextBytes, err := NewAES256CBCCrypter(iv, encryptionKey, false).Encrypt([]byte(actualHashedPassword))
	require.NoError(t, err)

	require.Equal(t, expectedCipherText, base64.StdEncoding.EncodeToString(cipherTextBytes))
}

func TestPKCS7(t *testing.T) {
	data := []byte("UIEAF8tALbXU451yQ4TMDa6D/sS4OLeW8xaphJztNjk=")
	paddedBytes, err := PKCS7Pad(data, aes.BlockSize)
	require.NoError(t, err)

	strippedBytes, err := PKCS7Strip(paddedBytes, aes.BlockSize)
	require.NoError(t, err)

	require.Equal(t, data, strippedBytes)
}

// write a test for AES256GCMEncrypt
func TestAES256GCM(t *testing.T) {
	iv, err := hex.DecodeString(commonIV12)
	require.NoError(t, err)

	encryptionKey, err := hex.DecodeString(commonKey)
	require.NoError(t, err)

	expectedCipherText := "Dr8M84rD8xD2qoZHS59XvpH3lQ9XPjQR86kCsQ=="
	result, err := AES256GCMEncrypt(AES256GCMParams{
		Plaintext: []byte(commonSecret),
		Key:       encryptionKey,
		Nonce:     iv,
	})
	require.NoError(t, err)
	require.Equal(t, expectedCipherText, base64.StdEncoding.EncodeToString(result.Ciphertext))

	plaintext, err := AES256GCMDecrypt(result)
	require.NoError(t, err)
	require.Equal(t, commonSecret, string(plaintext))
}
