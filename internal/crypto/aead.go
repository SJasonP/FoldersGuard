package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

type SealedObject struct {
	Algorithm  string
	Nonce      []byte
	Ciphertext []byte
}

func NewAES256GCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize256 {
		return nil, fmt.Errorf("AES-256-GCM key must be %d bytes", KeySize256)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	return aead, nil
}

func SealAES256GCM(key, plaintext, associatedData []byte) (SealedObject, error) {
	aead, err := NewAES256GCM(key)
	if err != nil {
		return SealedObject{}, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return SealedObject{}, fmt.Errorf("generate nonce: %w", err)
	}

	return SealedObject{
		Algorithm:  "AES-256-GCM",
		Nonce:      nonce,
		Ciphertext: aead.Seal(nil, nonce, plaintext, associatedData),
	}, nil
}

func OpenAES256GCM(key []byte, sealed SealedObject, associatedData []byte) ([]byte, error) {
	if sealed.Algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported sealed object algorithm %q", sealed.Algorithm)
	}
	if len(sealed.Nonce) == 0 {
		return nil, errors.New("nonce is required")
	}

	aead, err := NewAES256GCM(key)
	if err != nil {
		return nil, err
	}
	if len(sealed.Nonce) != aead.NonceSize() {
		return nil, fmt.Errorf("nonce must be %d bytes", aead.NonceSize())
	}

	plaintext, err := aead.Open(nil, sealed.Nonce, sealed.Ciphertext, associatedData)
	if err != nil {
		return nil, fmt.Errorf("open sealed object: %w", err)
	}
	return plaintext, nil
}
