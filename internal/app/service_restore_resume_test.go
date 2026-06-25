package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceDecryptProjectResumeSkipsExistingAndRestoresMissing(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	password := "test-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "a.txt"), []byte("alpha"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "b.txt"), []byte("bravo-content"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(filepath.Join(root, "data"))
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

	if _, err := service.DecryptProject(ctx, DecryptProjectInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		OutputRoot:    restored,
		SourceCleanup: SourceCleanupKeep,
	}); err != nil {
		t.Fatal(err)
	}

	base := filepath.Base(source)
	keep := filepath.Join(restored, base, "a.txt")
	missing := filepath.Join(restored, base, "b.txt")

	past := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(keep, past, past); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}

	// Resume restores the missing output and leaves the existing one untouched.
	// A non-empty output directory must not be rejected when resuming.
	if _, err := service.DecryptProject(ctx, DecryptProjectInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		OutputRoot:    restored,
		SourceCleanup: SourceCleanupKeep,
		Resume:        true,
	}); err != nil {
		t.Fatalf("resume decrypt: %v", err)
	}

	if got, err := os.ReadFile(missing); err != nil || string(got) != "bravo-content" {
		t.Fatalf("missing output not restored: %q err=%v", got, err)
	}
	info, err := os.Stat(keep)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().UTC().Truncate(time.Second).Equal(past) {
		t.Fatalf("existing output was rewritten on resume (mtime %s)", info.ModTime())
	}
}
