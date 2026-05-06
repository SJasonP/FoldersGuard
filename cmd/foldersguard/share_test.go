package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		"--out-database", shareDatabase,
		"--password-env", "FG_TEST_PASSWORD",
		"--share-password-env", "FG_SHARE_PASSWORD",
	}, nil, &shareOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, shareOutput.String(),
		"share_database="+shareDatabase+"\n",
		"items=2\n",
		"files=1\n",
		"folders=1\n",
		"password_protected=true\n",
	)

	copyShareContentFromOutput(t, shareOutput.String(), contentOutput, shareContent)

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
		"--out-database", shareDatabase,
		"--password-env", "FG_TEST_PASSWORD",
		"--no-share-password",
	}, nil, &shareOutput); err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, shareOutput.String(), "password_protected=false\n")

	copyShareContentFromOutput(t, shareOutput.String(), contentOutput, shareContent)

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

func copyShareContentFromOutput(t *testing.T, shareOutput, contentOutput, shareContent string) {
	t.Helper()
	locations := contentLocationsFromOutput(t, shareOutput)
	if len(locations) == 0 {
		t.Fatal("missing content_location output")
	}
	for sourceVisible, targetVisible := range locations {
		copySharePath(t, filepath.Join(contentOutput, filepath.FromSlash(sourceVisible)), filepath.Join(shareContent, filepath.FromSlash(targetVisible)))
	}
}

func contentLocationsFromOutput(t *testing.T, output string) map[string]string {
	t.Helper()
	locations := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "content_location ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			t.Fatalf("content_location line = %q", line)
		}
		source := strings.TrimPrefix(fields[1], "source=")
		target := strings.TrimPrefix(fields[2], "target=")
		if source == fields[1] || target == fields[2] {
			t.Fatalf("content_location line = %q", line)
		}
		locations[source] = target
	}
	return locations
}

func copySharePath(t *testing.T, source, target string) {
	t.Helper()
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(target, relative)
			if entry.IsDir() {
				return os.MkdirAll(targetPath, 0o755)
			}
			return copyShareFile(path, targetPath)
		}); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := copyShareFile(source, target); err != nil {
		t.Fatal(err)
	}
}

func copyShareFile(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		_ = output.Close()
		if !committed {
			_ = os.Remove(target)
		}
	}()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	committed = true
	return nil
}
