package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
