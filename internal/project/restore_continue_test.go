package project

import (
	"context"
	"path/filepath"
	"testing"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

func TestRestorerContinueOnErrorRecordsAndProceeds(t *testing.T) {
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
	if err := (Executor{OutputRoot: output}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	// Corrupt the encrypted object for b.txt so its decryption authentication
	// fails while the others restore correctly.
	visibleByItem := map[string]string{}
	for _, obj := range plan.StorageObjects {
		if obj.Type == model.StorageObjectTypeFile {
			visibleByItem[obj.ItemID.String()] = obj.VisiblePath
		}
	}
	var corruptedID string
	for _, file := range plan.Files {
		if filepath.Base(file.SourcePath) == "b.txt" {
			corruptedID = file.ID.String()
			encPath := filepath.Join(output, filepath.FromSlash(visibleByItem[file.ID.String()]))
			mustWrite(t, encPath, []byte("this is not a valid encrypted object at all"))
		}
	}

	restored := filepath.Join(root, "restored")
	var failed []string
	report, err := (Restorer{
		EncryptedRoot:   output,
		OutputRoot:      restored,
		ContinueOnError: true,
		OnFileError: func(file model.File, err error) {
			failed = append(failed, file.ID.String())
		},
	}).RestoreContentReport(context.Background(), plan)
	if err != nil {
		t.Fatalf("continue-on-error should not abort: %v", err)
	}
	if len(failed) != 1 || failed[0] != corruptedID {
		t.Fatalf("recorded failures = %v, want [%s]", failed, corruptedID)
	}

	base := filepath.Base(source)
	assertFile(t, filepath.Join(restored, base, "a.txt"), []byte("alpha"))
	assertFile(t, filepath.Join(restored, base, "c.txt"), []byte("charlie"))
	if report.DecryptedFiles != 2 {
		t.Fatalf("expected 2 files restored, got %d", report.DecryptedFiles)
	}
}

func TestRestorerAbortsOnErrorByDefault(t *testing.T) {
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
	if err := (Executor{OutputRoot: output}).EncryptContent(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	visibleByItem := map[string]string{}
	for _, obj := range plan.StorageObjects {
		if obj.Type == model.StorageObjectTypeFile {
			visibleByItem[obj.ItemID.String()] = obj.VisiblePath
		}
	}
	for _, file := range plan.Files {
		encPath := filepath.Join(output, filepath.FromSlash(visibleByItem[file.ID.String()]))
		mustWrite(t, encPath, []byte("corrupt"))
	}

	restored := filepath.Join(root, "restored")
	_, err = (Restorer{EncryptedRoot: output, OutputRoot: restored}).RestoreContentReport(context.Background(), plan)
	if err == nil {
		t.Fatal("expected the default mode to abort on the first error")
	}
}
