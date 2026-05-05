package project

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/content"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

func TestExecutorEncryptContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	mustMkdir(t, filepath.Join(source, "dir"))
	mustWrite(t, filepath.Join(source, "dir", "small.txt"), []byte("small"))
	mustWrite(t, filepath.Join(source, "large.txt"), []byte("large-content"))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 5}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}

	executor := Executor{OutputRoot: output}
	if err := executor.EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	for _, object := range plan.StorageObjects {
		path := filepath.Join(output, filepath.FromSlash(object.VisiblePath))
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("missing storage object %s: %v", object.VisiblePath, err)
		}
		switch object.Type {
		case model.StorageObjectTypeFolder:
			if !info.IsDir() {
				t.Fatalf("%s is not a directory", object.VisiblePath)
			}
		case model.StorageObjectTypeFile, model.StorageObjectTypePart:
			if info.IsDir() {
				t.Fatalf("%s is a directory", object.VisiblePath)
			}
		}
	}
}

func TestRestorerRestoreContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustMkdir(t, filepath.Join(source, "dir"))
	mustWrite(t, filepath.Join(source, "dir", "small.txt"), []byte("small"))
	mustWrite(t, filepath.Join(source, "large.txt"), []byte("large-content"))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 5}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := (Executor{OutputRoot: encrypted}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	if err := (Restorer{EncryptedRoot: encrypted, OutputRoot: restored}).RestoreContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(restored, filepath.Base(source), "dir", "small.txt"), []byte("small"))
	assertFile(t, filepath.Join(restored, filepath.Base(source), "large.txt"), []byte("large-content"))
}

func TestRestorerRejectsTamperedContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "file.txt"), []byte("secret"))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := (Executor{OutputRoot: encrypted}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	var objectPath string
	for _, object := range plan.StorageObjects {
		if object.Type == model.StorageObjectTypeFile {
			objectPath = filepath.Join(encrypted, filepath.FromSlash(object.VisiblePath))
			break
		}
	}
	data, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(objectPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	err = (Restorer{EncryptedRoot: encrypted, OutputRoot: restored}).RestoreContent(context.Background(), plan)
	if err == nil {
		t.Fatal("expected authentication failure")
	}
	if _, err := os.Stat(filepath.Join(restored, filepath.Base(source), "file.txt")); err == nil {
		t.Fatal("tampered file was restored")
	}
}

func TestRestorerRejectsPathLikeRealName(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustWrite(t, filepath.Join(source, "file.txt"), []byte("secret"))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := (Executor{OutputRoot: encrypted}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	plan.Items[0].RealName = "../escape.txt"

	err = (Restorer{EncryptedRoot: encrypted, OutputRoot: restored}).RestoreContent(context.Background(), plan)
	if err == nil {
		t.Fatal("expected path-like name rejection")
	}
	if _, err := os.Stat(filepath.Join(root, "escape.txt")); err == nil {
		t.Fatal("restore escaped output root")
	}
}

func TestSafeJoinRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	for _, relative := range []string{"../x", "/tmp/x", "a/../../x"} {
		if _, err := content.SafeJoin(root, relative); err == nil {
			t.Fatalf("expected %q to be rejected", relative)
		}
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}
