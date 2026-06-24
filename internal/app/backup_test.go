package app

import (
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/format"
)

// writeActiveDatabase writes a fake active project database with the given
// content. The backup engine copies file bytes, so a real SQLCipher database is
// not required to exercise it.
func writeActiveDatabase(t *testing.T, service Service, projectID string, content []byte) string {
	t.Helper()
	path, err := service.ActiveProjectDatabasePath(projectID)
	if err != nil {
		t.Fatalf("active database path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create projects dir: %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write active database: %v", err)
	}
	return path
}

func newTestService(t *testing.T) Service {
	t.Helper()
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}

func TestBackupProjectDatabaseCopiesBytes(t *testing.T) {
	service := newTestService(t)
	const projectID = "11111111-1111-1111-1111-111111111111"
	writeActiveDatabase(t, service, projectID, []byte("encrypted-db-v1"))

	path, err := service.backupProjectDatabase(projectID, BackupReasonApply)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if path == "" {
		t.Fatal("expected a backup path")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != "encrypted-db-v1" {
		t.Fatalf("backup content = %q, want %q", got, "encrypted-db-v1")
	}

	backups, err := service.ListProjectBackups(projectID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("backup count = %d, want 1", len(backups))
	}
	if backups[0].Reason != BackupReasonApply {
		t.Fatalf("reason = %q, want %q", backups[0].Reason, BackupReasonApply)
	}
	if backups[0].CreatedAt.IsZero() {
		t.Fatal("expected a parsed creation time")
	}
}

func TestBackupProjectDatabaseMissingActiveIsNoOp(t *testing.T) {
	service := newTestService(t)
	path, err := service.backupProjectDatabase("22222222-2222-2222-2222-222222222222", BackupReasonDelete)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if path != "" {
		t.Fatalf("expected no backup path, got %q", path)
	}
}

func TestProjectBackupsRetentionPrunes(t *testing.T) {
	service := newTestService(t)
	if _, err := service.SaveSettings(Settings{BackupRetention: 2}); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	const projectID = "33333333-3333-3333-3333-333333333333"

	for i := 0; i < 4; i++ {
		writeActiveDatabase(t, service, projectID, []byte{byte('a' + i)})
		if _, err := service.backupProjectDatabase(projectID, BackupReasonApply); err != nil {
			t.Fatalf("backup %d: %v", i, err)
		}
	}

	backups, err := service.ListProjectBackups(projectID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(backups) != 2 {
		t.Fatalf("retained backups = %d, want 2", len(backups))
	}
	// Newest first: the two most recent backups hold the last two contents.
	newest, err := os.ReadFile(backups[0].Path)
	if err != nil {
		t.Fatalf("read newest: %v", err)
	}
	if string(newest) != "d" {
		t.Fatalf("newest backup content = %q, want %q", newest, "d")
	}
}

func TestRestoreProjectBackupReplacesActive(t *testing.T) {
	service := newTestService(t)
	const projectID = "44444444-4444-4444-4444-444444444444"
	writeActiveDatabase(t, service, projectID, []byte("original"))

	if _, err := service.backupProjectDatabase(projectID, BackupReasonApply); err != nil {
		t.Fatalf("backup: %v", err)
	}
	backups, err := service.ListProjectBackups(projectID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("backup count = %d, want 1", len(backups))
	}

	// Mutate the active database, then restore from the backup.
	activePath := writeActiveDatabase(t, service, projectID, []byte("changed"))
	if _, err := service.RestoreProjectBackup(projectID, backups[0].ID, true); err != nil {
		t.Fatalf("restore: %v", err)
	}
	got, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("read active: %v", err)
	}
	if string(got) != "original" {
		t.Fatalf("restored content = %q, want %q", got, "original")
	}

	// Restoring takes a pre-restore backup of the changed database.
	after, err := service.ListProjectBackups(projectID)
	if err != nil {
		t.Fatalf("list after restore: %v", err)
	}
	if len(after) != 2 {
		t.Fatalf("backup count after restore = %d, want 2", len(after))
	}
}

func TestRestoreProjectBackupRequiresForceWhenActiveExists(t *testing.T) {
	service := newTestService(t)
	const projectID = "55555555-5555-5555-5555-555555555555"
	writeActiveDatabase(t, service, projectID, []byte("original"))
	if _, err := service.backupProjectDatabase(projectID, BackupReasonApply); err != nil {
		t.Fatalf("backup: %v", err)
	}
	backups, err := service.ListProjectBackups(projectID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, err := service.RestoreProjectBackup(projectID, backups[0].ID, false); err == nil {
		t.Fatal("expected restore without force to fail when active database exists")
	}
}

func TestRestoreProjectBackupRejectsUnknownBackup(t *testing.T) {
	service := newTestService(t)
	if _, err := service.RestoreProjectBackup("66666666-6666-6666-6666-666666666666", "nope", true); err == nil {
		t.Fatal("expected error for unknown backup id")
	}
}

func TestListProjectBackupsEmpty(t *testing.T) {
	service := newTestService(t)
	backups, err := service.ListProjectBackups("77777777-7777-7777-7777-777777777777")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(backups) != 0 {
		t.Fatalf("backup count = %d, want 0", len(backups))
	}
}

func TestProjectBackupFilesAreOwnerOnly(t *testing.T) {
	service := newTestService(t)
	const projectID = "88888888-8888-8888-8888-888888888888"
	writeActiveDatabase(t, service, projectID, []byte("secret"))
	path, err := service.backupProjectDatabase(projectID, BackupReasonManual)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("backup permissions = %o, want 600", perm)
	}
}

func TestBackupExtensionIsProjectExtension(t *testing.T) {
	service := newTestService(t)
	const projectID = "99999999-9999-9999-9999-999999999999"
	writeActiveDatabase(t, service, projectID, []byte("x"))
	path, err := service.backupProjectDatabase(projectID, BackupReasonApply)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if !format.IsProjectExtension(path) {
		t.Fatalf("backup path %q does not use the project extension", path)
	}
}
