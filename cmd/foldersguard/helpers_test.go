package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/db"
	"foldersguard/internal/storage"
)

func assertOutputContains(t *testing.T, output string, want ...string) {
	t.Helper()
	for _, expected := range want {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Fatalf("output = %q, want %q", output, expected)
		}
	}
}

func outputValue(t *testing.T, output, key string) string {
	t.Helper()
	prefix := key + "="
	for _, line := range bytes.Split([]byte(output), []byte("\n")) {
		if bytes.HasPrefix(line, []byte(prefix)) {
			return string(bytes.TrimPrefix(line, []byte(prefix)))
		}
	}
	t.Fatalf("output = %q, missing key %q", output, key)
	return ""
}

func tamperFirstEncryptedFile(t *testing.T, root string) {
	t.Helper()
	var encryptedFile string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && encryptedFile == "" {
			encryptedFile = path
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(encryptedFile)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(encryptedFile, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertProjectDatabaseOpens(t *testing.T, path, password string) {
	t.Helper()
	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       path,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		t.Fatal(err)
	}
	meta, err := store.Meta(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if meta["database_crypto_suite"] != "SQLCipher" {
		t.Fatalf("database_crypto_suite = %q, want SQLCipher", meta["database_crypto_suite"])
	}
}

func assertProjectDatabaseRejectsPassword(t *testing.T, path, password string) {
	t.Helper()
	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       path,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Meta(ctx); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}
