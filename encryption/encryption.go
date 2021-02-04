package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

var cipherSuite = "AES-256"
var keyLength = 32

type CipherSuite struct {
	Cipher string
	KeyLength int
}

func Describe() CipherSuite {
	return CipherSuite{
		Cipher:    cipherSuite,
		KeyLength: keyLength,
	}
}

func EncryptFile(key []byte, filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed reading file: %v", err)
	}

	encrypted := Encrypt(key, data)

	err = ioutil.WriteFile(filename + ".encrypted", encrypted, 0644)
	if err != nil {
		return fmt.Errorf("failed writing contents to file: %v", err)
	}
	return nil
}

func DecryptFile(key []byte, filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed reading file: %v", err)
	}

	decrypted, err := Decrypt(key, data)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	newFilename := strings.TrimSuffix(filename, ".encrypted")
	err = ioutil.WriteFile(newFilename, decrypted, 0644)
	if err != nil {
		return "", fmt.Errorf("failed writing contents to file: %v", err)
	}
	return newFilename, nil
}

func Encrypt(key []byte, data []byte) []byte {
	gcm := cipherBlock(key)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(fmt.Errorf("error randomizing nonce: %v", err))
	}

	return gcm.Seal(nonce, nonce, data, nil)
}

func Decrypt(key []byte, data []byte) ([]byte, error) {
	gcm := cipherBlock(key)
	nonce := data[:gcm.NonceSize()]
	data = data[gcm.NonceSize():]
	decrypted, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed decrypting data: %v", err)
	}

	return decrypted, nil
}

func GenerateKey() []byte {
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return key
}

func cipherBlock(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("failed creating cipher block: %v", err))
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(fmt.Errorf("failed creating GCM cipher block: %v", err))
	}

	return gcm
}