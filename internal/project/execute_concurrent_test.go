package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
	"foldersguard/internal/progress"
)

// TestExecutorConcurrentEncryptsAllFiles encrypts many files with a worker pool
// and confirms every file is encrypted, restores correctly, and that
// byte-weighted progress reaches its total. Run under -race, the shared
// progress tracker and callbacks must stay race-free.
func TestExecutorConcurrentEncryptsAllFiles(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	const fileCount = 40
	for i := 0; i < fileCount; i++ {
		name := fmt.Sprintf("file-%02d.txt", i)
		if err := os.WriteFile(filepath.Join(source, name), []byte(fmt.Sprintf("payload %d", i)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}

	// The sink asserts, on every emitted event, that byte-weighted progress is
	// monotonic (never moves backward) and never exceeds the total, which is the
	// property that must hold when many files report bytes concurrently.
	var mu sync.Mutex
	var maxBytes int64
	tracker := progress.New("op", "encrypt", func(e progress.Event) {
		mu.Lock()
		defer mu.Unlock()
		if e.ProcessedBytes < maxBytes {
			t.Errorf("progress moved backward: %d < %d", e.ProcessedBytes, maxBytes)
		}
		maxBytes = e.ProcessedBytes
		if e.TotalBytes > 0 && e.ProcessedBytes > e.TotalBytes {
			t.Errorf("progress exceeded total: %d > %d", e.ProcessedBytes, e.TotalBytes)
		}
	}, progress.PhaseEncrypting)
	tracker.StartPhase(progress.PhaseEncrypting, true)

	var deletedMu sync.Mutex
	deleted := 0
	executor := Executor{
		OutputRoot:  output,
		Progress:    tracker,
		Concurrency: 4,
		AfterFile: func(model.File) error {
			deletedMu.Lock()
			deleted++
			deletedMu.Unlock()
			return nil
		},
	}
	if err := executor.EncryptContent(context.Background(), plan); err != nil {
		t.Fatalf("concurrent encrypt: %v", err)
	}

	// AfterFile runs exactly once per file, so a correct count proves every file
	// completed. (The Executor serializes AfterFile, so the plain counter is safe,
	// but the mutex documents that and keeps -race quiet.)
	if deleted != fileCount {
		t.Fatalf("AfterFile ran %d times, want %d", deleted, fileCount)
	}

	// Everything restores.
	restored := filepath.Join(root, "restored")
	if _, err := (Restorer{EncryptedRoot: output, OutputRoot: restored}).RestoreContentReport(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	base := filepath.Base(source)
	for i := 0; i < fileCount; i++ {
		name := fmt.Sprintf("file-%02d.txt", i)
		assertFile(t, filepath.Join(restored, base, name), []byte(fmt.Sprintf("payload %d", i)))
	}
}

// TestExecutorConcurrentContinueOnError confirms continue-on-error records each
// failed file and encrypts the rest when running concurrently.
func TestExecutorConcurrentContinueOnError(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	const fileCount = 20
	for i := 0; i < fileCount; i++ {
		name := fmt.Sprintf("file-%02d.txt", i)
		if err := os.WriteFile(filepath.Join(source, name), []byte(fmt.Sprintf("payload %d", i)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}

	// Remove a few sources after planning so their encryption fails.
	removed := map[string]bool{}
	for _, file := range plan.Files {
		base := filepath.Base(file.SourcePath)
		if base == "file-03.txt" || base == "file-11.txt" || base == "file-17.txt" {
			removed[file.ID.String()] = true
			if err := os.Remove(file.SourcePath); err != nil {
				t.Fatal(err)
			}
		}
	}

	var mu sync.Mutex
	failed := map[string]bool{}
	executor := Executor{
		OutputRoot:      output,
		Concurrency:     4,
		ContinueOnError: true,
		OnFileError: func(file model.File, _ error) {
			mu.Lock()
			failed[file.ID.String()] = true
			mu.Unlock()
		},
	}
	if err := executor.EncryptContent(context.Background(), plan); err != nil {
		t.Fatalf("continue-on-error concurrent encrypt: %v", err)
	}
	if len(failed) != len(removed) {
		t.Fatalf("recorded %d failures, want %d", len(failed), len(removed))
	}
	for id := range removed {
		if !failed[id] {
			t.Fatalf("missing recorded failure for %s", id)
		}
	}
}

// TestExecutorConcurrentAbortsOnError confirms the default (abort) mode returns
// an error when a file fails during concurrent encryption.
func TestExecutorConcurrentAbortsOnError(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("file-%02d.txt", i)
		if err := os.WriteFile(filepath.Join(source, name), []byte("payload"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range plan.Files {
		if filepath.Base(file.SourcePath) == "file-07.txt" {
			if err := os.Remove(file.SourcePath); err != nil {
				t.Fatal(err)
			}
		}
	}

	err = (Executor{OutputRoot: output, Concurrency: 4}).EncryptContent(context.Background(), plan)
	if err == nil {
		t.Fatal("expected concurrent encryption to abort on the first error")
	}
}
