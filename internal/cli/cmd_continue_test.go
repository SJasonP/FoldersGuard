package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestDecryptContinueOnErrorReportsFailuresAndExitsNonZero(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentRoot := filepath.Join(root, "content")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(name+" content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")

	var encryptOut bytes.Buffer
	if err := RunWithIO("fg", []string{
		"encrypt", source,
		"--content-out", contentRoot,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &encryptOut); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	projectID := fieldValue(t, encryptOut.String(), "project_id")

	// Corrupt one encrypted file object so exactly one decryption fails.
	var encFiles []string
	if err := filepath.WalkDir(contentRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			encFiles = append(encFiles, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(encFiles) < 2 {
		t.Fatalf("expected at least 2 encrypted objects, got %d", len(encFiles))
	}
	sort.Strings(encFiles)
	if err := os.WriteFile(encFiles[0], []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}

	// The default mode aborts with a non-zero exit.
	if err := RunWithIO("fg", []string{
		"decrypt", projectID,
		"--content", contentRoot,
		"--out", filepath.Join(root, "restored-abort"),
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil); err == nil {
		t.Fatal("expected the default decrypt to abort on the corrupt object")
	}

	// Continue-on-error records the failure, restores the rest, and still exits
	// non-zero because an item failed.
	var out, errOut bytes.Buffer
	err := RunWithIOErr("fg", []string{
		"decrypt", projectID,
		"--content", contentRoot,
		"--out", filepath.Join(root, "restored-continue"),
		"--continue-on-error",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, &out, &errOut)
	if err == nil {
		t.Fatal("expected a non-zero exit when an item fails under continue-on-error")
	}
	if !strings.Contains(out.String(), "failed_files=1\n") {
		t.Fatalf("stdout = %q, want failed_files=1", out.String())
	}
	if !strings.Contains(out.String(), "restored_files=1\n") {
		t.Fatalf("stdout = %q, want restored_files=1", out.String())
	}
	if !strings.Contains(errOut.String(), "failed_file=") {
		t.Fatalf("stderr = %q, want a failed_file line", errOut.String())
	}
}

func fieldValue(t *testing.T, output, key string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, key+"=") {
			return strings.TrimPrefix(line, key+"=")
		}
	}
	t.Fatalf("output %q missing field %q", output, key)
	return ""
}
