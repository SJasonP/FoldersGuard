package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

func TestExecutorContinueOnErrorRecordsAndProceeds(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))
	mustWrite(t, filepath.Join(source, "b.txt"), []byte("bravo"))
	mustWrite(t, filepath.Join(source, "c.txt"), []byte("charlie"))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}

	// Make one source file unreadable by removing it after planning, so its
	// encryption fails while the others succeed.
	var removedID string
	for _, file := range plan.Files {
		if filepath.Base(file.SourcePath) == "b.txt" {
			removedID = file.ID.String()
			if err := os.Remove(file.SourcePath); err != nil {
				t.Fatal(err)
			}
		}
	}

	var failed []string
	executor := Executor{
		OutputRoot:      output,
		ContinueOnError: true,
		OnFileError: func(file model.File, err error) {
			failed = append(failed, file.ID.String())
		},
	}
	if err := executor.EncryptContent(context.Background(), plan); err != nil {
		t.Fatalf("continue-on-error should not abort: %v", err)
	}
	if len(failed) != 1 || failed[0] != removedID {
		t.Fatalf("recorded failures = %v, want [%s]", failed, removedID)
	}

	// The other files were encrypted and restore correctly.
	restored := filepath.Join(root, "restored")
	report, err := (Restorer{EncryptedRoot: output, OutputRoot: restored}).RestoreContentReport(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Base(source)
	assertFile(t, filepath.Join(restored, base, "a.txt"), []byte("alpha"))
	assertFile(t, filepath.Join(restored, base, "c.txt"), []byte("charlie"))
	if report.DecryptedFiles < 2 {
		t.Fatalf("expected at least 2 files restored, got %d", report.DecryptedFiles)
	}
}

func TestExecutorAbortsOnErrorByDefault(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	mustWrite(t, filepath.Join(source, "a.txt"), []byte("alpha"))
	mustWrite(t, filepath.Join(source, "b.txt"), []byte("bravo"))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range plan.Files {
		if filepath.Base(file.SourcePath) == "a.txt" {
			if err := os.Remove(file.SourcePath); err != nil {
				t.Fatal(err)
			}
		}
	}

	err = (Executor{OutputRoot: output}).EncryptContent(context.Background(), plan)
	if err == nil {
		t.Fatal("expected the default mode to abort on the first error")
	}
}
