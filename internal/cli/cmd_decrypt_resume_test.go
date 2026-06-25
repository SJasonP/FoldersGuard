package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDecryptResumeSkipsExistingAndRestoresMissing(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "a.txt"), []byte("alpha"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "b.txt"), []byte("bravo-content"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_PASSWORD", "test-password")

	var encryptOut bytes.Buffer
	if err := RunWithIO("fg", []string{
		"encrypt", source,
		"--content-out", contentOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_PASSWORD",
	}, nil, &encryptOut); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	projectID := parseProjectID(t, encryptOut.String())

	if err := RunWithIO("fg", []string{
		"decrypt", projectID,
		"--content", contentOutput,
		"--out", restored,
		"--password-env", "FG_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	base := filepath.Base(source)
	missing := filepath.Join(restored, base, "b.txt")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}

	// Resume accepts the non-empty output and restores only the missing file.
	if err := RunWithIO("fg", []string{
		"decrypt", projectID,
		"--content", contentOutput,
		"--out", restored,
		"--resume",
		"--password-env", "FG_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatalf("decrypt --resume: %v", err)
	}
	if got, err := os.ReadFile(missing); err != nil || string(got) != "bravo-content" {
		t.Fatalf("missing output not restored on resume: %q err=%v", got, err)
	}
}

func TestDecryptForceAndResumeAreMutuallyExclusive(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("FG_PASSWORD", "test-password")
	root := t.TempDir()
	content := filepath.Join(root, "encrypted")
	out := filepath.Join(root, "out")
	if err := os.MkdirAll(content, 0o755); err != nil {
		t.Fatal(err)
	}
	err := RunWithIO("fg", []string{
		"decrypt", "some-project",
		"--content", content,
		"--out", out,
		"--force",
		"--resume",
		"--password-env", "FG_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected --force and --resume to be mutually exclusive")
	}
}
