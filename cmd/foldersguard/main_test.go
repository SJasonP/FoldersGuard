package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/cli"
	"foldersguard/internal/db"
	"foldersguard/internal/storage"
)

func TestRunEncryptCreatesEncryptedContentAndDatabase(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	restoreOutput := filepath.Join(root, "restored")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello foldersguard")
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), plaintext, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	if err := cli.RunWithIO("foldersguard", []string{
		"encrypt",
		source,
		"--content-out", contentOutput,
		"--export", databaseOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(databaseOutput); err != nil {
		t.Fatalf("database output missing: %v", err)
	}
	databaseBytes, err := os.ReadFile(databaseOutput)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(databaseBytes, []byte("note.txt")) {
		t.Fatal("encrypted database contains plaintext filename")
	}
	if bytes.HasPrefix(databaseBytes, []byte("SQLite format 3\x00")) {
		t.Fatal("project database has plaintext SQLite header")
	}
	assertProjectDatabaseOpens(t, databaseOutput, "test-password")
	assertProjectDatabaseRejectsPassword(t, databaseOutput, "wrong-password")

	var encryptedFiles int
	err = filepath.WalkDir(contentOutput, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			encryptedFiles++
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Contains(data, plaintext) {
				t.Fatalf("encrypted content %s contains plaintext", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if encryptedFiles != 1 {
		t.Fatalf("encrypted files = %d, want 1", encryptedFiles)
	}

	var inspectOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"inspect",
		databaseOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &inspectOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, inspectOutput.String(),
		"database_type=project\n",
		"root_name=source\n",
		"files=1\n",
		"parts=0\n",
		"storage_objects=3\n",
	)

	if err := cli.RunWithIO("foldersguard", []string{
		"decrypt",
		databaseOutput,
		"--content", contentOutput,
		"--out", restoreOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(restoreOutput, "source", "docs", "note.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored, plaintext) {
		t.Fatalf("restored plaintext = %q, want %q", restored, plaintext)
	}

	t.Setenv("FG_WRONG_PASSWORD", "wrong-password")
	wrongPasswordOutput := filepath.Join(root, "wrong-password")
	err = cli.RunWithIO("foldersguard", []string{
		"decrypt",
		databaseOutput,
		"--content", contentOutput,
		"--out", wrongPasswordOutput,
		"--password-env", "FG_WRONG_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected wrong password to fail")
	}
	if _, statErr := os.Stat(wrongPasswordOutput); !os.IsNotExist(statErr) {
		t.Fatalf("wrong-password output stat error = %v, want not exist", statErr)
	}
}

func assertOutputContains(t *testing.T, output string, want ...string) {
	t.Helper()
	for _, expected := range want {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Fatalf("output = %q, want %q", output, expected)
		}
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

func TestRunEncryptRejectsOutputInsideSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	err := cli.RunWithIO("foldersguard", []string{
		"encrypt",
		source,
		"--content-out", filepath.Join(source, "content"),
		"--export", filepath.Join(root, "project.fg"),
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected output-inside-source error")
	}
}

func TestRunEncryptRejectsOutputEqualToSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	err := cli.RunWithIO("foldersguard", []string{
		"encrypt",
		source,
		"--content-out", source,
		"--export", filepath.Join(root, "project.fg"),
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected output-equals-source error")
	}
}
