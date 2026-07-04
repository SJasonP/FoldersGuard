package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceCreateProjectConcurrentRestores(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	dataDir := filepath.Join(root, "data")
	password := "test-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	const fileCount = 12
	for i := 0; i < fileCount; i++ {
		name := fmt.Sprintf("file-%02d.txt", i)
		if err := os.WriteFile(filepath.Join(source, name), []byte(fmt.Sprintf("payload %d", i)), 0o600); err != nil {
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
		Concurrency:   4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.EncryptedFiles != fileCount || created.FailedFiles != 0 {
		t.Fatalf("create result = %+v", created)
	}

	decrypted, err := service.DecryptProject(ctx, DecryptProjectInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		OutputRoot:    restored,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}
	if decrypted.DecryptedFiles != fileCount {
		t.Fatalf("decrypted %d files, want %d", decrypted.DecryptedFiles, fileCount)
	}
	base := filepath.Base(source)
	for i := 0; i < fileCount; i++ {
		name := fmt.Sprintf("file-%02d.txt", i)
		data, err := os.ReadFile(filepath.Join(restored, base, name))
		if err != nil || string(data) != fmt.Sprintf("payload %d", i) {
			t.Fatalf("restored %s = %q, err = %v", name, data, err)
		}
	}
}

func TestDefaultEncryptionConcurrencyIsPositive(t *testing.T) {
	if got := DefaultEncryptionConcurrency(); got < 1 {
		t.Fatalf("DefaultEncryptionConcurrency() = %d, want >= 1", got)
	}
}

func TestResolveEncryptionConcurrencyOverride(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.resolveEncryptionConcurrency(3)
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Fatalf("override 3 resolved to %d", got)
	}
	got, err = service.resolveEncryptionConcurrency(0)
	if err != nil {
		t.Fatal(err)
	}
	if got < 1 {
		t.Fatalf("zero override resolved to %d, want >= 1", got)
	}
}
