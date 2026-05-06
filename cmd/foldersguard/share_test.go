package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/cli"
)

func TestRunShareCreatesRestorableShare(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	shareContent := filepath.Join(root, "share-content")
	shareDatabase := filepath.Join(root, "share.fgs")
	shareRestore := filepath.Join(root, "share-restored")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	sharedPlaintext := []byte("shared")
	privatePlaintext := []byte("private")
	if err := os.WriteFile(filepath.Join(source, "docs", "shared.txt"), sharedPlaintext, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "private", "secret.txt"), privatePlaintext, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	t.Setenv("FG_SHARE_PASSWORD", "share-password")
	var encryptOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"encrypt",
		source,
		"--content-out", contentOutput,
		"--export", databaseOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &encryptOutput); err != nil {
		t.Fatal(err)
	}
	projectID := outputValue(t, encryptOutput.String(), "project_id")

	var shareOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"share",
		projectID,
		"source/docs",
		"--content", contentOutput,
		"--out-content", shareContent,
		"--out-database", shareDatabase,
		"--password-env", "FG_TEST_PASSWORD",
		"--share-password-env", "FG_SHARE_PASSWORD",
	}, nil, &shareOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, shareOutput.String(),
		"share_database="+shareDatabase+"\n",
		"share_content="+shareContent+"\n",
		"items=2\n",
		"files=1\n",
		"folders=1\n",
		"password_protected=true\n",
	)

	var verifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		shareDatabase,
		"--content", shareContent,
		"--password-env", "FG_SHARE_PASSWORD",
	}, nil, &verifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, verifyOutput.String(),
		"extra_objects=0\n",
		"status=ok\n",
	)

	if err := cli.RunWithIO("foldersguard", []string{
		"decrypt",
		shareDatabase,
		"--content", shareContent,
		"--out", shareRestore,
		"--password-env", "FG_SHARE_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(shareRestore, "docs", "shared.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored, sharedPlaintext) {
		t.Fatalf("restored plaintext = %q, want %q", restored, sharedPlaintext)
	}
	if _, err := os.Stat(filepath.Join(shareRestore, "private", "secret.txt")); !os.IsNotExist(err) {
		t.Fatalf("private file stat error = %v, want not exist", err)
	}
}

func TestRunShareCreatesUnprotectedShare(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	shareContent := filepath.Join(root, "share-content")
	shareDatabase := filepath.Join(root, "share.fgs")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("shared"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	var encryptOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"encrypt",
		source,
		"--content-out", contentOutput,
		"--export", databaseOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &encryptOutput); err != nil {
		t.Fatal(err)
	}
	projectID := outputValue(t, encryptOutput.String(), "project_id")

	var shareOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"share",
		projectID,
		"source/note.txt",
		"--content", contentOutput,
		"--out-content", shareContent,
		"--out-database", shareDatabase,
		"--password-env", "FG_TEST_PASSWORD",
		"--no-share-password",
	}, nil, &shareOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, shareOutput.String(), "password_protected=false\n")

	var verifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		shareDatabase,
		"--content", shareContent,
	}, nil, &verifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, verifyOutput.String(), "status=ok\n")
}
