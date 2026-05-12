package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/fswalk"
)

func TestVerifierVerifyContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
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

	report, err := (Verifier{EncryptedRoot: encrypted}).VerifyContent(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("report = %+v, want ok", report)
	}
	if report.CheckedObjects != len(plan.StorageObjects) {
		t.Fatalf("checked objects = %d, want %d", report.CheckedObjects, len(plan.StorageObjects))
	}

	var encryptedFile string
	err = filepath.WalkDir(encrypted, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && encryptedFile == "" {
			encryptedFile = path
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(encryptedFile)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(encryptedFile, data, 0o600); err != nil {
		t.Fatal(err)
	}

	report, err = (Verifier{EncryptedRoot: encrypted}).VerifyContent(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if report.TamperedObjects != 1 {
		t.Fatalf("tampered objects = %d, want 1", report.TamperedObjects)
	}
}

func TestVerifierExtraObjectsDoNotFailVerification(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.MkdirAll(filepath.Join(encrypted, "extra-shared-folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(encrypted, "extra-shared-folder", "extra-object"), []byte("other encrypted content"), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := (Verifier{EncryptedRoot: encrypted}).VerifyContent(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if report.ExtraObjects != 2 {
		t.Fatalf("extra objects = %d, want 2", report.ExtraObjects)
	}
	if len(report.ExtraPaths) != 2 || report.ExtraPaths[0] != "extra-shared-folder" || report.ExtraPaths[1] != "extra-shared-folder/extra-object" {
		t.Fatalf("extra paths = %#v", report.ExtraPaths)
	}
	if !report.OK() {
		t.Fatalf("report = %+v, want ok despite extra objects", report)
	}
}

func TestVerifierReportsProblemPaths(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "missing.txt"), []byte("missing"), 0o644); err != nil {
		t.Fatal(err)
	}

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

	visiblePaths := visiblePathsByItem(plan)
	pathsBySource := make(map[string]string)
	for _, file := range plan.Files {
		pathsBySource[filepath.Base(file.SourcePath)] = visiblePaths[file.ID.String()]
	}
	tamperedPath := pathsBySource["note.txt"]
	missingPath := pathsBySource["missing.txt"]

	filePath := filepath.Join(encrypted, filepath.FromSlash(tamperedPath))
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(encrypted, filepath.FromSlash(missingPath))); err != nil {
		t.Fatal(err)
	}

	report, err := (Verifier{EncryptedRoot: encrypted}).VerifyContent(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if report.MissingObjects != 1 || len(report.MissingPaths) != 1 || report.MissingPaths[0] != filepath.Clean(filepath.FromSlash(missingPath)) {
		t.Fatalf("missing details = %+v, want %s", report, missingPath)
	}
	if report.TamperedObjects != 1 || len(report.TamperedPaths) != 1 || report.TamperedPaths[0] != filepath.Clean(filepath.FromSlash(tamperedPath)) {
		t.Fatalf("tampered details = %+v, want %s", report, tamperedPath)
	}
}

func TestVerifierIgnoresSystemMetadataFiles(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

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
	for _, name := range []string{".DS_Store", "._temporary", "Thumbs.db", "desktop.ini"} {
		if err := os.WriteFile(filepath.Join(encrypted, name), []byte("metadata"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(encrypted, ".fseventsd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(encrypted, ".fseventsd", "event-log"), []byte("metadata"), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := (Verifier{EncryptedRoot: encrypted}).VerifyContent(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if report.ExtraObjects != 0 {
		t.Fatalf("extra objects = %d, want 0 for system metadata", report.ExtraObjects)
	}
	if !report.OK() {
		t.Fatalf("report = %+v, want ok", report)
	}
}
