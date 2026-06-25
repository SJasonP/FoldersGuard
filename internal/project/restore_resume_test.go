package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRestorerResumeSkipsExistingOutputs(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))
	mustWrite(t, filepath.Join(source, "b.txt"), []byte("bravo-content"))

	plan := planAndEncrypt(t, source, encrypted)
	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	base := filepath.Base(source)
	outputs := []string{
		filepath.Join(restored, base, "a.txt"),
		filepath.Join(restored, base, "b.txt"),
	}
	past := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, path := range outputs {
		if err := os.Chtimes(path, past, past); err != nil {
			t.Fatal(err)
		}
	}

	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored, Resume: true}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	for _, path := range outputs {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().UTC().Truncate(time.Second).Equal(past) {
			t.Fatalf("output %s was rewritten on resume (mtime %s)", path, info.ModTime())
		}
	}
}

func TestRestorerResumeRestoresMissingOutput(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))
	mustWrite(t, filepath.Join(source, "b.txt"), []byte("bravo-content"))

	plan := planAndEncrypt(t, source, encrypted)
	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	base := filepath.Base(source)
	missing := filepath.Join(restored, base, "a.txt")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}

	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored, Resume: true}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	assertFile(t, missing, []byte("alpha"))
}

func TestRestorerResumeRestoresWrongSizeOutput(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))

	plan := planAndEncrypt(t, source, encrypted)
	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	base := filepath.Base(source)
	output := filepath.Join(restored, base, "a.txt")
	// A truncated, wrong-size output is not trusted and is restored again.
	if err := os.WriteFile(output, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored, Resume: true}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	assertFile(t, output, []byte("alpha"))
}
