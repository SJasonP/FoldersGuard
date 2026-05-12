package app

import (
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/storage"
)

func TestApplyStorageContentOperations(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "content")
	stagingRoot := filepath.Join(root, "staging")
	if err := os.MkdirAll(filepath.Join(stagingRoot, "new-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagingRoot, "new-folder", "file.dat"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(contentRoot, "old-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentRoot, "old-folder", "file.dat"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(contentRoot, "delete-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentRoot, "delete-folder", "file.dat"), []byte("delete"), 0o600); err != nil {
		t.Fatal(err)
	}

	applied, err := ApplyStorageContentOperations([]storage.ContentOperation{
		{Type: "upload", SourcePath: "new-folder", TargetPath: "uploaded/new-folder"},
		{Type: "move", SourcePath: "old-folder", TargetPath: "moved/old-folder"},
		{Type: "delete", TargetPath: "delete-folder"},
	}, ContentOperationApplyOptions{
		ContentRoot: contentRoot,
		StagingRoot: stagingRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 3 {
		t.Fatalf("applied operations = %+v", applied)
	}
	assertExists(t, filepath.Join(contentRoot, "uploaded", "new-folder", "file.dat"))
	assertMissing(t, filepath.Join(stagingRoot, "new-folder"))
	assertExists(t, filepath.Join(contentRoot, "moved", "old-folder", "file.dat"))
	assertMissing(t, filepath.Join(contentRoot, "old-folder"))
	assertMissing(t, filepath.Join(contentRoot, "delete-folder"))
}

func TestApplyStorageContentOperationsRejectsUnsafePath(t *testing.T) {
	_, err := ApplyStorageContentOperations([]storage.ContentOperation{{
		Type:       "move",
		SourcePath: "../outside",
		TargetPath: "target",
	}}, ContentOperationApplyOptions{ContentRoot: t.TempDir()})
	if err == nil {
		t.Fatal("expected unsafe path to be rejected")
	}
}

func TestApplyStorageContentOperationsWithCommitRollsBackMoveOnCommitFailure(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "content")
	if err := os.MkdirAll(filepath.Join(contentRoot, "old-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentRoot, "old-folder", "file.dat"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ApplyStorageContentOperationsWithCommit([]storage.ContentOperation{{
		Type:       "move",
		SourcePath: "old-folder",
		TargetPath: "moved/old-folder",
	}}, ContentOperationApplyOptions{
		ContentRoot: contentRoot,
	}, func() error {
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("expected commit failure")
	}
	assertExists(t, filepath.Join(contentRoot, "old-folder", "file.dat"))
	assertMissing(t, filepath.Join(contentRoot, "moved", "old-folder"))
}

func TestApplyStorageContentOperationsWithCommitRollsBackDeleteOnCommitFailure(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "content")
	if err := os.MkdirAll(filepath.Join(contentRoot, "delete-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentRoot, "delete-folder", "file.dat"), []byte("delete"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ApplyStorageContentOperationsWithCommit([]storage.ContentOperation{{
		Type:       "delete",
		TargetPath: "delete-folder",
	}}, ContentOperationApplyOptions{
		ContentRoot: contentRoot,
	}, func() error {
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("expected commit failure")
	}
	assertExists(t, filepath.Join(contentRoot, "delete-folder", "file.dat"))
}

func TestApplyStorageContentOperationsWithCommitRollsBackUploadOnCommitFailure(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "content")
	stagingRoot := filepath.Join(root, "staging")
	if err := os.MkdirAll(filepath.Join(stagingRoot, "new-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagingRoot, "new-folder", "file.dat"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ApplyStorageContentOperationsWithCommit([]storage.ContentOperation{{
		Type:       "upload",
		SourcePath: "new-folder",
		TargetPath: "uploaded/new-folder",
	}}, ContentOperationApplyOptions{
		ContentRoot: contentRoot,
		StagingRoot: stagingRoot,
	}, func() error {
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("expected commit failure")
	}
	assertMissing(t, filepath.Join(contentRoot, "uploaded", "new-folder"))
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be missing, stat error = %v", path, err)
	}
}
