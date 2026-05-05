package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/cli"
)

func TestRunEncryptCreatesEncryptedContentAndDatabase(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	databaseExport := filepath.Join(root, "exported.fg")
	renamedRestoreOutput := filepath.Join(root, "renamed-restored")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello foldersguard")
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

	var verifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		databaseOutput,
		"--content", contentOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &verifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, verifyOutput.String(),
		"checked_objects=3\n",
		"missing_objects=0\n",
		"tampered_objects=0\n",
		"extra_objects=0\n",
		"status=ok\n",
	)

	var renameOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"rename",
		databaseOutput,
		"source/docs/note.txt",
		"renamed.txt",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &renameOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, renameOutput.String(),
		"old_name=note.txt\n",
		"new_name=renamed.txt\n",
		"content_operations=0\n",
	)

	if err := cli.RunWithIO("foldersguard", []string{
		"decrypt",
		databaseOutput,
		"--content", contentOutput,
		"--out", renamedRestoreOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(renamedRestoreOutput, "source", "docs", "renamed.txt"))
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

	var inspectRenamedOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"inspect",
		databaseOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &inspectRenamedOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, inspectRenamedOutput.String(), "files=1\n")

	tamperFirstEncryptedFile(t, contentOutput)
	var tamperedVerifyOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"verify",
		databaseOutput,
		"--content", contentOutput,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &tamperedVerifyOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, tamperedVerifyOutput.String(),
		"tampered_objects=1\n",
		"status=failed\n",
	)

	var exportOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"export",
		projectID,
		"--out", databaseExport,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &exportOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, exportOutput.String(),
		"project_id="+projectID+"\n",
		"database_output="+databaseExport+"\n",
	)

	t.Setenv("HOME", filepath.Join(root, "import-home"))
	var importOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"import",
		databaseExport,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &importOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, importOutput.String(),
		"project_id="+projectID+"\n",
		"imported=true\n",
	)

	var importedInspectOutput bytes.Buffer
	if err := cli.RunWithIO("foldersguard", []string{
		"inspect",
		projectID,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &importedInspectOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, importedInspectOutput.String(),
		"project_id="+projectID+"\n",
		"database_type=project\n",
	)
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
