package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/crypto/argon2"
)

const (
	KeySize256 = 32
	SaltSize   = 16
)

type Argon2idParams struct {
	Time        uint32
	MemoryKiB   uint32
	Parallelism uint8
	KeyLen      uint32
}

func DefaultArgon2idParams() Argon2idParams {
	parallelism := runtime.NumCPU()
	if parallelism < 1 {
		parallelism = 1
	}
	if parallelism > 4 {
		parallelism = 4
	}

	return Argon2idParams{
		Time:        3,
		MemoryKiB:   256 * 1024,
		Parallelism: uint8(parallelism),
		KeyLen:      KeySize256,
	}
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}

func DeriveKey(password string, salt []byte, params Argon2idParams) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password is required")
	}
	if len(salt) < SaltSize {
		return nil, fmt.Errorf("salt must be at least %d bytes", SaltSize)
	}
	if params.Time == 0 || params.MemoryKiB == 0 || params.Parallelism == 0 || params.KeyLen == 0 {
		return nil, errors.New("invalid Argon2id parameters")
	}

	key := argon2.IDKey([]byte(password), salt, params.Time, params.MemoryKiB, params.Parallelism, params.KeyLen)
	return key, nil
}
