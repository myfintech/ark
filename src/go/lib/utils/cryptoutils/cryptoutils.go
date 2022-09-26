package cryptoutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/utils"
)

// RandomIV generates a random initialization vector of length n
func RandomIV(length int) ([]byte, error) {
	iv := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return iv, err
	}
	return iv, nil
}

// SHA256Sum hashes the input and returns a string in the specified encoding format
func SHA256Sum(data, format string) (string, error) {
	result := sha256.Sum256([]byte(data))
	return utils.EncodeBytesToString(result[:], format)
}

// SHA1Sum hashes the input and returns a string in the specified encoding format
func SHA1Sum(data, format string) (string, error) {
	result := sha1.Sum([]byte(data))
	return utils.EncodeBytesToString(result[:], format)
}

// CompareHashes returns true if the provided hashes are equal
func CompareHashes(expected, recieved hash.Hash) bool {
	return bytes.Equal(expected.Sum(nil), recieved.Sum(nil))
}

// MD5Sum hashes the input and returns a string in the specified encoding format
func MD5Sum(data, format string) (string, error) {
	result := md5.Sum([]byte(data))
	return utils.EncodeBytesToString(result[:], format)
}

// PKCS7Strip remove pkcs7 padding
func PKCS7Strip(data []byte, blockSize int) ([]byte, error) {
	length := len(data)

	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}

	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}

	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)

	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}

	return data[:length-padLen], nil
}

// PKCS7Pad add pkcs7 padding
func PKCS7Pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize < 0 || blockSize > 256 {
		return nil, fmt.Errorf("pkcs7: Invalid block size %d", blockSize)
	} else {
		padLen := 16 - len(data)%blockSize
		padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
		return append(data, padding...), nil
	}
}

// HMAC256Sum hashes the input and returns a string in the specified encoding format
func HMAC256Sum(message, key, format string) (string, error) {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return utils.EncodeBytesToString(mac.Sum(nil), format)
}

// HMACOptions ...
type HMACOptions struct {
	Message    string
	Key        string
	KeyFormat  string
	HashFormat string
}

// HMAC256V2 hashes the input and returns a string in the specified encoding format

// AES256CBCCrypter hashes the input and returns a string in the specified encoding format
type AES256CBCCrypter struct {
	Key                 []byte
	IV                  []byte
	BlockPaddingEnabled bool
}

// NewAES256CBCCrypter returns an AES256CBCCrypter
func NewAES256CBCCrypter(iv, key []byte, disablePadding bool) *AES256CBCCrypter {
	return &AES256CBCCrypter{
		IV:                  iv,
		Key:                 key,
		BlockPaddingEnabled: disablePadding != true,
	}
}

// Encrypt accepts plainText bytes and returns encrypts them using aes-256-cbc returning cipherText bytes
func (crypter *AES256CBCCrypter) Encrypt(plainText []byte) ([]byte, error) {
	// CBC (Cipher Block Chaining) requires each block of bytes to be an integral multiple of it's block size.
	// We pad the cipherText with empty bytes up to the next multiple of aes.BlockSize(16)
	if crypter.BlockPaddingEnabled {
		paddedBytes, err := PKCS7Pad(plainText, aes.BlockSize)
		if err != nil {
			return []byte{}, errors.Wrap(err, "failed to apply default block padding")
		}
		plainText = paddedBytes
	}

	cipherText := make([]byte, len(plainText))
	cipherBlock, err := aes.NewCipher(crypter.Key)

	if err != nil {
		return []byte{}, err
	}

	copy(cipherText, plainText)

	cbcEncrypter := cipher.NewCBCEncrypter(cipherBlock, crypter.IV)

	if len(cipherText)%cbcEncrypter.BlockSize() != 0 {
		return []byte{}, errors.Errorf("The length of the secret(%d) must be divisible by the block size(%d)", len(plainText), cbcEncrypter.BlockSize())
	}

	cbcEncrypter.CryptBlocks(cipherText, cipherText)

	return cipherText, nil
}

// Decrypt decrypts cip
func (crypter *AES256CBCCrypter) Decrypt(cipherText []byte) ([]byte, error) {
	plainText := make([]byte, len(cipherText))
	cipherBlock, err := aes.NewCipher(crypter.Key)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to create cipherBlock")
	}

	cbcDecrypter := cipher.NewCBCDecrypter(cipherBlock, crypter.IV)
	cbcDecrypter.CryptBlocks(plainText, cipherText)

	if crypter.BlockPaddingEnabled {
		strippedBytes, err := PKCS7Strip(plainText, aes.BlockSize)
		if err != nil {
			return []byte{}, errors.Wrap(err, "failed to strip default block padding")
		}
		plainText = strippedBytes
	}

	return plainText, nil
}

func AES256CBCEncrypt(iv []byte, plainText, key, encoding string) (string, error) {
	// CBC (Cipher Block Chaining) requires each block of bytes to be an integral multiple of it's block size.
	// We pad the cipherTextBytes with empty bytes up to the next multiple of aes.BlockSize(16)
	plainTextBytes, err := PKCS7Pad([]byte(plainText), aes.BlockSize)
	if err != nil {
		return "", err
	}

	cipherTextBytes := make([]byte, len(plainTextBytes))

	cipherBlock, err := aes.NewCipher([]byte(key))

	if err != nil {
		return "", err
	}

	copy(cipherTextBytes, plainTextBytes)

	// return "", errors.Errorf("ctb: %d ptb: %d", len(cipherTextBytes), len(plainTextBytes))

	cbcEncrypter := cipher.NewCBCEncrypter(cipherBlock, iv)

	if len(cipherTextBytes)%cbcEncrypter.BlockSize() != 0 {
		return "", errors.Errorf("The length of the secret(%d) must be divisible by the blocksize(%d)", len(plainTextBytes), cbcEncrypter.BlockSize())
	}

	cbcEncrypter.CryptBlocks(cipherTextBytes, cipherTextBytes)

	return utils.EncodeBytesToString(cipherTextBytes, encoding)
}

func AES256CBCDecrypt(cipherText, key string) (string, error) {
	cipherTextBytes := []byte(cipherText)

	block, err := aes.NewCipher([]byte(key))

	if err != nil {
		return "", err
	}

	if len(cipherTextBytes) < aes.BlockSize {
		return "", errors.New("cipherText must be greater than aes.BlockSize (16)")
	}

	iv := cipherTextBytes[:aes.BlockSize]

	// Remove the IV from the cipherTextBytes
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes, cipherTextBytes)

	return string(cipherTextBytes), nil
}

func AES256GCMNounce(gcm cipher.AEAD) (nonce []byte, err error) {
	// Create a new nonce
	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	return
}

type AES256GCMParams struct {
	Key            []byte
	Nonce          []byte
	Ciphertext     []byte
	Plaintext      []byte
	AdditionalData []byte
}

func AES256GCMEncrypt(params AES256GCMParams) (result AES256GCMParams, err error) {
	// AES-256 is 32 bytes
	if len(params.Key) != 32 {
		err = errors.Errorf("key must be 32 bytes, got %d", len(params.Key))
		return
	}

	// Create the AES cipher
	block, err := aes.NewCipher(params.Key)
	if err != nil {
		return
	}

	// GCM is a Galois/Counter Mode that uses a 256 bit key
	// GCM is a block cipher mode that provides both confidentiality and authenticity
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	if params.Nonce == nil {
		if params.Nonce, err = AES256GCMNounce(gcm); err != nil {
			return
		}
	}

	// Encrypt
	params.Ciphertext = gcm.Seal(nil, params.Nonce, params.Plaintext, params.AdditionalData)

	result = params
	return
}

func AES256GCMDecrypt(params AES256GCMParams) (plaintext []byte, err error) {
	// AES-256 is 32 bytes
	if len(params.Key) != 32 {
		err = errors.Errorf("key must be 32 bytes, got %d", len(params.Key))
		return
	}

	// Create the AES cipher
	block, err := aes.NewCipher(params.Key)
	if err != nil {
		return nil, err
	}

	// GCM is a Galois/Counter Mode that uses a 256 bit key
	// GCM is a block cipher mode that provides both confidentiality and authenticity
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err = gcm.Open(nil, params.Nonce, params.Ciphertext, params.AdditionalData)
	return
}
