package app

import (
	"context"
	"fmt"
	"os"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
)

// ChangeProjectPassword changes the password of an active project database by
// re-keying it. Internal content keys are unchanged and no encrypted content is
// rewritten.
func (s Service) ChangeProjectPassword(ctx context.Context, projectID, oldPassword, newPassword string) error {
	path, err := s.ActiveProjectDatabasePath(projectID)
	if err != nil {
		return err
	}
	return s.changeDatabasePassword(ctx, path, oldPassword, newPassword, projectID, "project")
}

// ChangeSharePassword changes the password of a share database by re-keying it.
// A share password change protects only future copies of the share database;
// copies already distributed are independent and unaffected.
func (s Service) ChangeSharePassword(ctx context.Context, sharePath, oldPassword, newPassword string) error {
	if !format.IsSetExtension(sharePath) {
		return fmt.Errorf("share database must use %s extension", format.SetExtension)
	}
	if err := ValidateDatabasePath(sharePath); err != nil {
		return err
	}
	return s.changeDatabasePassword(ctx, sharePath, oldPassword, newPassword, "", "share")
}

// changeDatabasePassword re-keys a database safely: it verifies the current
// password, backs up the project database, re-keys a copy, confirms the copy
// opens under the new password, then atomically replaces the live database. An
// interruption never leaves the user without a working database.
func (s Service) changeDatabasePassword(ctx context.Context, path, oldPassword, newPassword, backupProjectID, databaseType string) error {
	if oldPassword == "" {
		return fmt.Errorf("current password is required")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}

	_, meta, err := ReadDatabase(ctx, db.Config{Path: path, DriverName: db.SQLCipherDriver, Password: oldPassword})
	if err != nil {
		return err
	}
	if got := meta["database_type"]; got != databaseType {
		return fmt.Errorf("database type = %q, want %s", got, databaseType)
	}

	// Back up the project database before changing it. Share databases are
	// external files, and the copy-rekey-verify-swap below is itself crash-safe.
	if backupProjectID != "" {
		if _, err := s.backupProjectDatabase(backupProjectID, BackupReasonRekey); err != nil {
			return err
		}
	}

	temp := path + ".rekey.tmp"
	_ = os.Remove(temp)
	if err := CopyFile(path, temp); err != nil {
		return fmt.Errorf("stage rekey copy: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(temp)
		}
	}()

	if err := db.Rekey(ctx, db.Config{Path: temp, DriverName: db.SQLCipherDriver, Password: oldPassword}, newPassword); err != nil {
		return err
	}
	if _, _, err := ReadDatabase(ctx, db.Config{Path: temp, DriverName: db.SQLCipherDriver, Password: newPassword}); err != nil {
		return fmt.Errorf("verify rekeyed database: %w", err)
	}
	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("commit rekeyed database: %w", err)
	}
	committed = true
	return nil
}
