package fswalk

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"foldersguard/internal/fsmeta"
	"foldersguard/internal/noise"
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
	if result.Root.Metadata.Mode == 0 {
		t.Fatal("root metadata mode was not captured")
	}
	if result.Root.Metadata.ModTime.IsZero() {
		t.Fatal("root metadata modification time was not captured")
	}
	if !hasCapability(result.Root.Metadata.Capabilities, fsmeta.CapabilityMode) {
		t.Fatal("root metadata missing mode capability")
	}
	if !hasCapability(result.Root.Metadata.Capabilities, fsmeta.CapabilityModTime) {
		t.Fatal("root metadata missing modification time capability")
	}
}

func TestScanTopFolderCapturesRestorableMetadata(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	mustWrite(t, file, []byte("x"))
	wantMod := time.Date(2023, 2, 3, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes(file, wantMod, wantMod); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := ScanTopFolder(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(result.Entries))
	}
	metadata := result.Entries[0].Metadata
	if metadata.Mode&0o777 != 0o600 {
		t.Fatalf("metadata mode = %o, want 600", metadata.Mode&0o777)
	}
	if !sameFilesystemSecond(metadata.ModTime, wantMod) {
		t.Fatalf("metadata mod time = %s, want %s", metadata.ModTime, wantMod)
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

func TestScanTopFolderNoiseHandling(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".DS_Store"), []byte("finder metadata"))
	mustWrite(t, filepath.Join(root, "file.txt"), []byte("content"))

	ignored, err := ScanTopFolderWithNoiseMode(root, noise.ModeIgnoreEverywhere)
	if err != nil {
		t.Fatal(err)
	}
	if hasEntry(ignored, ".DS_Store") {
		t.Fatal(".DS_Store was included with ignore everywhere")
	}
	if !hasEntry(ignored, "file.txt") {
		t.Fatal("file.txt missing with ignore everywhere")
	}

	included, err := ScanTopFolderWithNoiseMode(root, noise.ModeDoNotIgnore)
	if err != nil {
		t.Fatal(err)
	}
	if !hasEntry(included, ".DS_Store") {
		t.Fatal(".DS_Store missing with do not ignore")
	}
}

func hasEntry(result ScanResult, path string) bool {
	for _, entry := range result.Entries {
		if entry.RootRelativePath == path {
			return true
		}
	}
	return false
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

func hasCapability(capabilities []string, want string) bool {
	for _, capability := range capabilities {
		if capability == want {
			return true
		}
	}
	return false
}

func sameFilesystemSecond(got, want time.Time) bool {
	return got.UTC().Truncate(time.Second).Equal(want.UTC().Truncate(time.Second))
}
