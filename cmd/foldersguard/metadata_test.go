package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/cli"
)

func TestRunRemoveDeletesMetadataAndEncryptedContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("remove me"), 0o644); err != nil {
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

	var removeOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"remove",
		projectID,
		"source/note.txt",
		"--content", contentOutput,
		"--force",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &removeOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, removeOutput.String(),
		"operations=1\n",
		"operation=delete target=",
	)

	var inspectOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"inspect",
		projectID,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &inspectOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, inspectOutput.String(), "files=0\n", "storage_objects=1\n")

	var verifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		projectID,
		"--content", contentOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &verifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, verifyOutput.String(),
		"checked_objects=1\n",
		"extra_objects=0\n",
		"status=ok\n",
	)
}

func TestRunMoveUpdatesMetadataAndEncryptedContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	restoreOutput := filepath.Join(root, "restored")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("move me")
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), plaintext, 0o644); err != nil {
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

	var moveOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"move",
		projectID,
		"source/docs/note.txt",
		"source/archive",
		"--content", contentOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &moveOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, moveOutput.String(),
		"operations=1\n",
		"operation=move source=",
		" target=",
	)

	var verifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		projectID,
		"--content", contentOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &verifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, verifyOutput.String(),
		"checked_objects=4\n",
		"extra_objects=0\n",
		"status=ok\n",
	)

	if err := cli.RunWithIO("foldersguard", []string{
		"decrypt",
		projectID,
		"--content", contentOutput,
		"--out", restoreOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(restoreOutput, "source", "archive", "note.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored, plaintext) {
		t.Fatalf("restored plaintext = %q, want %q", restored, plaintext)
	}
}

func TestRunAddUpdatesMetadataAndEncryptedContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	addSource := filepath.Join(root, "new.txt")
	contentOutput := filepath.Join(root, "content")
	stagingOutput := filepath.Join(root, "staging")
	databaseOutput := filepath.Join(root, "project.fg")
	restoreOutput := filepath.Join(root, "restored")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	addedPlaintext := []byte("new content")
	if err := os.WriteFile(addSource, addedPlaintext, 0o644); err != nil {
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

	var addOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"add",
		projectID,
		addSource,
		"source/docs",
		"--staging-content", stagingOutput,
		"--content", contentOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &addOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, addOutput.String(),
		"staging_content="+stagingOutput+"\n",
		"operations=1\n",
		"operation=upload source=",
		" target=",
	)

	var verifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		projectID,
		"--content", contentOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &verifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, verifyOutput.String(),
		"checked_objects=4\n",
		"extra_objects=0\n",
		"status=ok\n",
	)

	if err := cli.RunWithIO("foldersguard", []string{
		"decrypt",
		projectID,
		"--content", contentOutput,
		"--out", restoreOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(restoreOutput, "source", "docs", "new.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored, addedPlaintext) {
		t.Fatalf("restored plaintext = %q, want %q", restored, addedPlaintext)
	}
}
