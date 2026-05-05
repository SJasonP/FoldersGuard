package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	sqlcipher "github.com/mutecomm/go-sqlcipher/v4"
	_ "modernc.org/sqlite"
)

const (
	PlainDriver     = "sqlite"
	SQLCipherDriver = "sqlite3"
)

type Config struct {
	Path       string
	DriverName string
	Password   string
}

func OpenProject(ctx context.Context, config Config) (*sql.DB, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.Path) == "" {
		return nil, errors.New("database path is required")
	}

	driverName := config.DriverName
	if driverName == "" {
		driverName = SQLCipherDriver
	}

	switch driverName {
	case SQLCipherDriver:
		return openSQLCipher(ctx, config.Path, config.Password)
	case PlainDriver:
		return openPlain(ctx, config.Path)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", driverName)
	}
}

func openPlain(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open(PlainDriver, path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}
	return db, nil
}

func openSQLCipher(ctx context.Context, path, password string) (*sql.DB, error) {
	if password == "" {
		return nil, errors.New("database password is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	dsn := sqlcipherDSN(path, password)
	db, err := sql.Open(SQLCipherDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLCipher database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping SQLCipher database: %w", err)
	}
	return db, nil
}

func sqlcipherDSN(path, password string) string {
	values := url.Values{}
	values.Set("_pragma_key", escapeSQLCipherPragmaString(password))
	values.Set("_pragma_cipher_page_size", "4096")
	values.Set("_foreign_keys", "on")
	values.Set("_busy_timeout", "5000")
	values.Set("_journal_mode", "DELETE")
	values.Set("_secure_delete", "on")
	return path + "?" + values.Encode()
}

func escapeSQLCipherPragmaString(value string) string {
	return strings.ReplaceAll(value, `"`, `""`)
}

var _ = sqlcipher.Version
