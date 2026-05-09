//go:build !cgo

package db

import "errors"

func SQLCipherAvailable() error {
	return errors.New("SQLCipher requires CGO; rebuild FoldersGuard with CGO_ENABLED=1 and a target-platform C compiler")
}
