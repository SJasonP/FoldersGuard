package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRekeyChangesPassword(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "project.fg")

	database, err := OpenProject(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "alpha"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE secret (value TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `INSERT INTO secret (value) VALUES ('kept')`); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Rekey(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "alpha"}, "beta"); err != nil {
		t.Fatalf("rekey: %v", err)
	}

	if _, err := OpenProject(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "alpha"}); err == nil {
		t.Fatal("expected the old password to fail after rekey")
	}

	reopened, err := OpenProject(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "beta"})
	if err != nil {
		t.Fatalf("open with new password: %v", err)
	}
	defer reopened.Close()
	var value string
	if err := reopened.QueryRowContext(ctx, `SELECT value FROM secret`).Scan(&value); err != nil {
		t.Fatalf("read row after rekey: %v", err)
	}
	if value != "kept" {
		t.Fatalf("row value = %q, want %q", value, "kept")
	}
}

func TestRekeyRejectsWrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "project.fg")
	database, err := OpenProject(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "alpha"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE secret (value TEXT)`); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Rekey(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "wrong"}, "beta"); err == nil {
		t.Fatal("expected rekey with the wrong current password to fail")
	}
}

func TestRekeyRequiresNewPassword(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "project.fg")
	if err := Rekey(ctx, Config{Path: path, DriverName: SQLCipherDriver, Password: "alpha"}, ""); err == nil {
		t.Fatal("expected rekey with an empty new password to fail")
	}
}
