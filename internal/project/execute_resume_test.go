package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

func planAndEncrypt(t *testing.T, source, output string) model.PlannedProject {
	t.Helper()
	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 5}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := (Executor{OutputRoot: output}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	return plan
}

func contentObjectPaths(output string, plan model.PlannedProject) []string {
	var paths []string
	for _, object := range plan.StorageObjects {
		if object.Type == model.StorageObjectTypeFile || object.Type == model.StorageObjectTypePart {
			paths = append(paths, filepath.Join(output, filepath.FromSlash(object.VisiblePath)))
		}
	}
	return paths
}

func TestExecutorResumeSkipsVerifiedObjects(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))
	mustWrite(t, filepath.Join(source, "b.txt"), []byte("bravo-content"))

	plan := planAndEncrypt(t, source, output)

	// Mark every encrypted object with a known past mtime; a skipped file keeps
	// it, a re-encrypted file overwrites it.
	past := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	objectPaths := contentObjectPaths(output, plan)
	for _, path := range objectPaths {
		if err := os.Chtimes(path, past, past); err != nil {
			t.Fatal(err)
		}
	}

	if err := (Executor{OutputRoot: output, Resume: true, ResumeVerify: true}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	for _, path := range objectPaths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().UTC().Truncate(time.Second).Equal(past) {
			t.Fatalf("object %s was rewritten on resume (mtime %s)", path, info.ModTime())
		}
	}
}

func TestExecutorResumeRewritesMissingObject(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))
	mustWrite(t, filepath.Join(source, "b.txt"), []byte("bravo-content"))

	plan := planAndEncrypt(t, source, output)
	objectPaths := contentObjectPaths(output, plan)
	if err := os.Remove(objectPaths[0]); err != nil {
		t.Fatal(err)
	}

	if err := (Executor{OutputRoot: output, Resume: true, ResumeVerify: true}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(objectPaths[0]); err != nil {
		t.Fatalf("missing object was not recreated on resume: %v", err)
	}

	// The whole project still restores correctly.
	if err := (Restorer{EncryptedRoot: output, OutputRoot: restored}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(restored, filepath.Base(source), "a.txt"), []byte("alpha"))
	assertFile(t, filepath.Join(restored, filepath.Base(source), "b.txt"), []byte("bravo-content"))
}

func TestExecutorResumeVerifyRewritesCorruptObject(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))

	plan := planAndEncrypt(t, source, output)
	objectPaths := contentObjectPaths(output, plan)

	data, err := os.ReadFile(objectPaths[0])
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(objectPaths[0], data, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := (Executor{OutputRoot: output, Resume: true, ResumeVerify: true}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	// The corrupt object was re-encrypted, so restore succeeds.
	if err := (Restorer{EncryptedRoot: output, OutputRoot: restored}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatalf("restore after resume re-encrypt: %v", err)
	}
	assertFile(t, filepath.Join(restored, filepath.Base(source), "a.txt"), []byte("alpha"))
}

func TestExecutorResumePresenceOnlyTrustsExisting(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))

	plan := planAndEncrypt(t, source, output)
	objectPaths := contentObjectPaths(output, plan)

	data, err := os.ReadFile(objectPaths[0])
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(objectPaths[0], data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Presence-only resume trusts the existing object and does not re-encrypt it,
	// so the corruption remains.
	if err := (Executor{OutputRoot: output, Resume: true, ResumeVerify: false}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(objectPaths[0])
	if err != nil {
		t.Fatal(err)
	}
	if after[len(after)-1] != data[len(data)-1] {
		t.Fatal("presence-only resume unexpectedly rewrote the object")
	}
}
