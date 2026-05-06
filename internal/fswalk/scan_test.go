package fswalk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanTopFolder(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "dir"))
	mustWrite(t, filepath.Join(root, "dir", "file.txt"), []byte("hello"))
	mustWrite(t, filepath.Join(root, "root.txt"), []byte("root"))
	mustHardlink(t, filepath.Join(root, "root.txt"), filepath.Join(root, "hardlink.txt"))

	if err := os.Symlink(filepath.Join(root, "root.txt"), filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}

	result, err := ScanTopFolder(root)
	if err != nil {
		t.Fatal(err)
	}

	entries := map[string]EntryType{}
	for _, entry := range result.Entries {
		entries[entry.RootRelativePath] = entry.Type
	}

	if entries["dir"] != EntryTypeFolder {
		t.Fatalf("dir type = %q, want folder", entries["dir"])
	}
	if entries["dir/file.txt"] != EntryTypeFile {
		t.Fatalf("dir/file.txt type = %q, want file", entries["dir/file.txt"])
	}
	if entries["root.txt"] != EntryTypeFile {
		t.Fatalf("root.txt type = %q, want file", entries["root.txt"])
	}
	if entries["hardlink.txt"] != EntryTypeFile {
		t.Fatalf("hardlink.txt type = %q, want file", entries["hardlink.txt"])
	}
	if _, ok := entries["link.txt"]; ok {
		t.Fatal("symlink was included, want ignored")
	}
}

func TestScanTopFolderRejectsFileRoot(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	mustWrite(t, file, []byte("x"))

	if _, err := ScanTopFolder(file); err == nil {
		t.Fatal("expected file root rejection")
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

func mustHardlink(t *testing.T, oldname, newname string) {
	t.Helper()
	if err := os.Link(oldname, newname); err != nil {
		t.Fatal(err)
	}
}
