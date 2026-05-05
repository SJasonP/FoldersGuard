package crypto

import (
	"crypto/rand"
	"fmt"
)

func GenerateKey256() ([]byte, error) {
	key := make([]byte, KeySize256)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}
