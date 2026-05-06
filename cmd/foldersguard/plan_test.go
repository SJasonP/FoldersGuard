package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/cli"
)

func TestRunPlanMetadataCommands(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	addSource := filepath.Join(root, "new.txt")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	stagingOutput := filepath.Join(root, "staging")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("note"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(addSource, []byte("new"), 0o644); err != nil {
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
		"plan",
		"add",
		projectID,
		addSource,
		"source/docs",
		"--staging-content", stagingOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &addOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, addOutput.String(), "operations=1\n", "operation=upload source=", " target=")
	if _, err := os.Stat(stagingOutput); !os.IsNotExist(err) {
		t.Fatalf("staging stat error = %v, want not exist", err)
	}

	var moveOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"plan",
		"move",
		projectID,
		"source/docs/note.txt",
		"source/archive",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &moveOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, moveOutput.String(), "operations=1\n", "operation=move source=", " target=")

	var removeOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"plan",
		"remove",
		projectID,
		"source/docs/note.txt",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &removeOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, removeOutput.String(), "operations=1\n", "operation=delete target=")

	var inspectOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"inspect",
		projectID,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &inspectOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, inspectOutput.String(), "files=1\n")
}
