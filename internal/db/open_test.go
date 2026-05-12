package db

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenProjectSQLCipherRequiresPassword(t *testing.T) {
	_, err := OpenProject(context.Background(), Config{
		Path:       filepath.Join(t.TempDir(), "project.fg"),
		DriverName: SQLCipherDriver,
	})
	if err == nil {
		t.Fatal("expected password error")
	}
}

func TestOpenProjectSQLCipherAcceptsQuotedPassword(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "project.fg")
	database, err := OpenProject(ctx, Config{
		Path:       path,
		DriverName: SQLCipherDriver,
		Password:   `pass"word`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE secret (value TEXT)`); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenProject(ctx, Config{
		Path:       path,
		DriverName: SQLCipherDriver,
		Password:   `pass"word`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if _, err := reopened.ExecContext(ctx, `SELECT * FROM secret`); err != nil {
		t.Fatal(err)
	}
}

func TestOpenProjectSQLCipherRejectsWrongPasswordAtOpen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "project.fg")
	database, err := OpenProject(ctx, Config{
		Path:       path,
		DriverName: SQLCipherDriver,
		Password:   "correct",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE secret (value TEXT)`); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = OpenProject(ctx, Config{
		Path:       path,
		DriverName: SQLCipherDriver,
		Password:   "wrong",
	})
	if err == nil {
		t.Fatal("expected wrong password to fail during open")
	}
	if !errors.Is(err, ErrInvalidDatabasePassword) {
		t.Fatalf("error = %v, want ErrInvalidDatabasePassword", err)
	}
}

func TestOpenProjectSQLCipherCreatesEncryptedDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "project.fg")
	database, err := OpenProject(ctx, Config{
		Path:       path,
		DriverName: SQLCipherDriver,
		Password:   "password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE secret (value TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `INSERT INTO secret (value) VALUES ('hidden-name')`); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(data, []byte("SQLite format 3\x00")) {
		t.Fatal("SQLCipher database has plaintext SQLite header")
	}
	if bytes.Contains(data, []byte("hidden-name")) {
		t.Fatal("SQLCipher database contains plaintext row")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("database permissions = %o, want 600", got)
	}

	reopened, err := OpenProject(ctx, Config{
		Path:       path,
		DriverName: SQLCipherDriver,
		Password:   "password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	var value string
	if err := reopened.QueryRowContext(ctx, `SELECT value FROM secret`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "hidden-name" {
		t.Fatalf("value = %q, want hidden-name", value)
	}
}
