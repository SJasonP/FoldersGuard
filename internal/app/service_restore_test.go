package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceDecryptProject(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	restored := filepath.Join(root, "restored")
	dataDir := filepath.Join(root, "data")
	password := "test-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
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
	if decrypted.ProjectID != created.ProjectID || decrypted.DecryptedFiles != 1 || decrypted.RestoredFolders != 2 {
		t.Fatalf("decrypt project result = %+v", decrypted)
	}
	if data, err := os.ReadFile(filepath.Join(restored, "source", "docs", "note.txt")); err != nil || string(data) != "hello" {
		t.Fatalf("restored file data = %q, err = %v", data, err)
	}
}
