package project

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/progress"
)

// TestProgressReachesTotalAcrossLifecycle confirms that encrypt, verify, and
// restore each report byte progress that reaches the planned total, with the
// final terminal event marking the work fully processed.
func TestProgressReachesTotalAcrossLifecycle(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	mustMkdir(t, filepath.Join(source, "dir"))
	mustWrite(t, filepath.Join(source, "dir", "small.txt"), []byte("small payload"))
	mustWrite(t, filepath.Join(source, "large.bin"), bytes.Repeat([]byte("x"), 50_000))

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	// A small max part size forces split files, exercising part-level progress.
	plan, err := Planner{MaxPartSize: 16_000}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}

	var wantBytes int64
	for _, file := range plan.Files {
		wantBytes += file.OriginalSize
	}

	run := func(name string, fn func(tracker *progress.Tracker) error) {
		var last progress.Event
		tracker := progress.New("op", name, func(e progress.Event) { last = e })
		tracker.Begin()
		if err := fn(tracker); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		tracker.Finish(nil, false)
		if last.State != string(progress.StateCompleted) {
			t.Fatalf("%s final state = %q, want completed", name, last.State)
		}
		if last.ProcessedBytes != last.TotalBytes {
			t.Fatalf("%s processed %d of %d bytes", name, last.ProcessedBytes, last.TotalBytes)
		}
		if last.TotalBytes != wantBytes {
			t.Fatalf("%s total bytes = %d, want %d", name, last.TotalBytes, wantBytes)
		}
	}

	run(progress.PhaseEncrypting, func(tracker *progress.Tracker) error {
		tracker.StartPhase(progress.PhaseEncrypting, true)
		return Executor{OutputRoot: encrypted, Progress: tracker}.EncryptContent(context.Background(), plan)
	})
	run(progress.PhaseVerifying, func(tracker *progress.Tracker) error {
		tracker.StartPhase(progress.PhaseVerifying, true)
		_, err := Verifier{EncryptedRoot: encrypted, Progress: tracker}.VerifyContent(context.Background(), plan)
		return err
	})
	run(progress.PhaseDecrypting, func(tracker *progress.Tracker) error {
		tracker.StartPhase(progress.PhaseDecrypting, true)
		_, err := Restorer{EncryptedRoot: encrypted, OutputRoot: restored, Progress: tracker}.RestoreContentReport(context.Background(), plan)
		return err
	})
}
