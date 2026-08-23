package project

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	wantFileTime := time.Date(2022, 3, 4, 5, 6, 7, 0, time.UTC)
	wantDirTime := time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes(filepath.Join(source, "dir", "small.txt"), wantFileTime, wantFileTime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(source, "dir", "small.txt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(source, "dir"), wantDirTime, wantDirTime); err != nil {
		t.Fatal(err)
	}

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
	assertMetadata(t, filepath.Join(restored, filepath.Base(source), "dir", "small.txt"), 0o600, wantFileTime)
	assertMetadata(t, filepath.Join(restored, filepath.Base(source), "dir"), 0, wantDirTime)
}

func TestRestorerRestoresRecognizedSubset(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustMkdir(t, filepath.Join(source, "docs"))
	mustWrite(t, filepath.Join(source, "docs", "note.txt"), []byte("note"))
	mustWrite(t, filepath.Join(source, "other.txt"), []byte("other"))

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

	subsetRoot := encryptedPathForRealPath(t, encrypted, plan, "source/docs")
	report, err := (Restorer{EncryptedRoot: subsetRoot, OutputRoot: restored}).RestoreContentReport(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}

	if report.DecryptedFiles != 1 || report.RestoredFolders != 2 || report.SkippedFolders != 0 {
		t.Fatalf("restore report = %+v", report)
	}
	assertFile(t, filepath.Join(restored, filepath.Base(source), "docs", "note.txt"), []byte("note"))
	if _, err := os.Stat(filepath.Join(restored, filepath.Base(source), "other.txt")); !os.IsNotExist(err) {
		t.Fatalf("other file stat error = %v, want not exist", err)
	}
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

func TestSplitIncrementalCleanupReleasesAndRestoresByPart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	plaintext := []byte("abcdefghijklmnopqrstuvwxyz")
	mustWrite(t, filepath.Join(source, "large.bin"), plaintext)
	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := (Planner{MaxPartSize: 8}).Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parts) < 2 {
		t.Fatalf("parts = %d, want split file", len(plan.Parts))
	}

	var truncatedAt []int64
	encrypted := filepath.Join(root, "encrypted")
	err = (Executor{
		OutputRoot:   encrypted,
		ReverseParts: true,
		AfterPart: func(file model.File, part model.Part) error {
			truncatedAt = append(truncatedAt, part.Offset)
			return os.Truncate(file.SourcePath, part.Offset)
		},
		AfterFile: func(file model.File) error { return os.Remove(file.SourcePath) },
	}).EncryptContent(ctx, plan)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(truncatedAt); i++ {
		if truncatedAt[i] >= truncatedAt[i-1] {
			t.Fatalf("truncate offsets = %v, want strictly descending", truncatedAt)
		}
	}

	deletedParts := 0
	restored := filepath.Join(root, "restored")
	_, err = (Restorer{
		EncryptedRoot:      encrypted,
		OutputRoot:         restored,
		IncrementalCleanup: true,
		AfterPart: func(part RestoredPart) error {
			deletedParts++
			return os.Remove(part.EncryptedAbsPath)
		},
	}).RestoreContentReport(ctx, plan)
	if err != nil {
		t.Fatal(err)
	}
	if deletedParts != len(plan.Parts) {
		t.Fatalf("deleted parts = %d, want %d", deletedParts, len(plan.Parts))
	}
	assertFile(t, filepath.Join(restored, filepath.Base(source), "large.bin"), plaintext)
}

func encryptedPathForRealPath(t *testing.T, encryptedRoot string, plan model.PlannedProject, realPath string) string {
	t.Helper()
	logicalPaths, err := logicalRealPaths(plan)
	if err != nil {
		t.Fatal(err)
	}
	visiblePaths := visiblePathsByItem(plan)
	for itemID, logicalPath := range logicalPaths {
		if logicalPath != realPath {
			continue
		}
		visiblePath := visiblePaths[itemID]
		if visiblePath == "" {
			t.Fatalf("visible path not found for %s", realPath)
		}
		return filepath.Join(encryptedRoot, filepath.FromSlash(visiblePath))
	}
	t.Fatalf("real path %s not found", realPath)
	return ""
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

func TestExecutorRejectsFolderOutsideOutputRoot(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "output")
	plan := model.PlannedProject{StorageObjects: []model.StorageObject{{
		Type:        model.StorageObjectTypeFolder,
		VisiblePath: "../escaped-folder",
	}}}
	if err := (Executor{OutputRoot: output}).createFolders(context.Background(), plan); err == nil {
		t.Fatal("expected unsafe folder visible path to be rejected")
	}
	if _, err := os.Stat(filepath.Join(root, "escaped-folder")); !os.IsNotExist(err) {
		t.Fatalf("folder outside output root was created: %v", err)
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

func assertMetadata(t *testing.T, path string, wantPerm os.FileMode, wantModTime time.Time) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if wantPerm != 0 && info.Mode().Perm() != wantPerm {
		t.Fatalf("%s mode = %o, want %o", path, info.Mode().Perm(), wantPerm)
	}
	if !info.ModTime().UTC().Truncate(time.Second).Equal(wantModTime.UTC().Truncate(time.Second)) {
		t.Fatalf("%s mod time = %s, want %s", path, info.ModTime(), wantModTime)
	}
}
