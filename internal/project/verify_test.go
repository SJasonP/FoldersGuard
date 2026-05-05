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
