package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureOperationSpaceRequiresFullOutputByDefault(t *testing.T) {
	original := inspectDiskSpace
	t.Cleanup(func() { inspectDiskSpace = original })
	inspectDiskSpace = func(string) (diskSpaceInfo, error) {
		return diskSpaceInfo{Available: diskSpaceReserve + 99, VolumeID: "disk"}, nil
	}
	err := ensureOperationSpace("/source", "/output", []int64{60, 40}, []int64{60, 40}, false)
	if !errors.Is(err, ErrInsufficientDiskSpace) {
		t.Fatalf("error = %v, want insufficient disk space", err)
	}
}

func TestCreateProjectRejectsInsufficientSpaceBeforeModifyingOutput(t *testing.T) {
	original := inspectDiskSpace
	t.Cleanup(func() { inspectDiskSpace = original })
	inspectDiskSpace = func(string) (diskSpaceInfo, error) {
		return diskSpaceInfo{Available: 0, VolumeID: "disk"}, nil
	}
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "source.txt"), []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(output, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.CreateProject(context.Background(), CreateProjectInput{
		SourcePath: source, ContentOutput: output, Password: "password", Force: true,
	})
	if !errors.Is(err, ErrInsufficientDiskSpace) {
		t.Fatalf("error = %v, want insufficient disk space", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("output was modified before preflight failure: %v", err)
	}
}

func TestEnsureOperationSpaceUsesIncrementalPeakOnSameVolume(t *testing.T) {
	original := inspectDiskSpace
	t.Cleanup(func() { inspectDiskSpace = original })
	inspectDiskSpace = func(string) (diskSpaceInfo, error) {
		return diskSpaceInfo{Available: diskSpaceReserve + 65, VolumeID: "disk"}, nil
	}
	if err := ensureOperationSpace("/source", "/output", []int64{60, 40}, []int64{60, 40}, true); err != nil {
		t.Fatalf("incremental space check: %v", err)
	}
}

func TestEnsureOperationSpaceCannotReuseSpaceAcrossVolumes(t *testing.T) {
	original := inspectDiskSpace
	t.Cleanup(func() { inspectDiskSpace = original })
	calls := 0
	inspectDiskSpace = func(string) (diskSpaceInfo, error) {
		calls++
		volume := "output"
		if calls == 2 {
			volume = "source"
		}
		return diskSpaceInfo{Available: diskSpaceReserve + 65, VolumeID: volume}, nil
	}
	err := ensureOperationSpace("/source", "/output", []int64{60, 40}, []int64{60, 40}, true)
	if !errors.Is(err, ErrInsufficientDiskSpace) {
		t.Fatalf("error = %v, want insufficient disk space", err)
	}
}

func TestIncrementalPeakIncludesAccumulatedEncryptionOverhead(t *testing.T) {
	got := incrementalPeakSize([]int64{110, 110}, []int64{100, 100})
	if got != 120 {
		t.Fatalf("peak = %d, want 120", got)
	}
}
