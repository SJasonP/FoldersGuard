package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestChangeProjectPasswordRekeysWithoutTouchingContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      "old-pass",
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := service.ChangeProjectPassword(ctx, created.ProjectID, "old-pass", "new-pass"); err != nil {
		t.Fatalf("change password: %v", err)
	}

	// The old password no longer opens the database.
	if _, err := service.Inspect(ctx, DatabaseOpen{ProjectRef: created.ProjectID, Password: "old-pass"}); err == nil {
		t.Fatal("expected the old password to fail after the change")
	}
	// The new password opens it, and the content still verifies, proving the
	// content keys were untouched.
	if _, err := service.Inspect(ctx, DatabaseOpen{ProjectRef: created.ProjectID, Password: "new-pass"}); err != nil {
		t.Fatalf("inspect with new password: %v", err)
	}
	verify, err := service.Verify(ctx, DatabaseOpen{ProjectRef: created.ProjectID, Password: "new-pass"}, encrypted)
	if err != nil {
		t.Fatalf("verify after rekey: %v", err)
	}
	if verify.Status != "ok" || verify.MissingObjects != 0 || verify.TamperedObjects != 0 {
		t.Fatalf("verify after rekey = %+v", verify)
	}

	// A rekey backup was taken before the change.
	backups, err := service.ListProjectBackups(created.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	foundRekeyBackup := false
	for _, backup := range backups {
		if backup.Reason == BackupReasonRekey {
			foundRekeyBackup = true
		}
	}
	if !foundRekeyBackup {
		t.Fatalf("expected a rekey backup, got %+v", backups)
	}
}

func TestChangeProjectPasswordRejectsWrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      "old-pass",
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := service.ChangeProjectPassword(ctx, created.ProjectID, "wrong-pass", "new-pass"); err == nil {
		t.Fatal("expected the wrong current password to be rejected")
	}
	// The original password still works after a rejected change.
	if _, err := service.Inspect(ctx, DatabaseOpen{ProjectRef: created.ProjectID, Password: "old-pass"}); err != nil {
		t.Fatalf("original password should still work: %v", err)
	}
}

func TestChangeProjectPasswordRequiresNewPassword(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ChangeProjectPassword(ctx, "11111111-1111-1111-1111-111111111111", "old", ""); err == nil {
		t.Fatal("expected an empty new password to be rejected")
	}
}
