package app

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestServiceDecryptProjectContinueOnError(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "test-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(name+" content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.FailedFiles != 0 || len(created.Failures) != 0 {
		t.Fatalf("clean create reported failures: %+v", created)
	}

	// Corrupt one encrypted file object so exactly one decryption fails.
	var encFiles []string
	if err := filepath.WalkDir(encrypted, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			encFiles = append(encFiles, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(encFiles) < 2 {
		t.Fatalf("expected at least 2 encrypted objects, got %d", len(encFiles))
	}
	sort.Strings(encFiles)
	if err := os.WriteFile(encFiles[0], []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}

	// The default mode aborts on the corrupt object.
	if _, err := service.DecryptProject(ctx, DecryptProjectInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		OutputRoot:    filepath.Join(root, "restored-abort"),
		SourceCleanup: SourceCleanupKeep,
	}); err == nil {
		t.Fatal("expected the default decrypt to abort on the corrupt object")
	}

	// Continue-on-error records the failed object and restores the rest.
	result, err := service.DecryptProject(ctx, DecryptProjectInput{
		ProjectID:       created.ProjectID,
		Password:        password,
		EncryptedRoot:   encrypted,
		OutputRoot:      filepath.Join(root, "restored-continue"),
		SourceCleanup:   SourceCleanupKeep,
		FailureHandling: FailureHandlingContinue,
	})
	if err != nil {
		t.Fatalf("continue-on-error decrypt: %v", err)
	}
	if result.FailedEncryptedFiles != 1 || len(result.Failures) != 1 {
		t.Fatalf("continue result = %+v", result)
	}
	if result.Failures[0].FileID == "" {
		t.Fatalf("failure is missing the visible file id: %+v", result.Failures[0])
	}
	if result.DecryptedFiles != 1 {
		t.Fatalf("expected 1 file restored, got %d", result.DecryptedFiles)
	}
}
