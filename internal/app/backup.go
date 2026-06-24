package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"foldersguard/internal/format"
)

// DefaultBackupRetention is the number of database backups kept per project when
// the retention setting is unset.
const DefaultBackupRetention = 10

// Backup reasons describe why a database backup was taken. They are recorded in
// the backup file name and shown when listing backups.
const (
	BackupReasonApply   = "apply"
	BackupReasonDelete  = "delete"
	BackupReasonRekey   = "rekey"
	BackupReasonRestore = "restore"
	BackupReasonManual  = "manual"
)

// backupTimeLayout is a fixed-width, lexicographically sortable UTC timestamp
// used in backup file names so that sorting by name sorts by time.
const backupTimeLayout = "20060102T150405.000000000"

// backupSeparator separates the timestamp and reason in a backup id. It is a
// character that never appears in the timestamp or in a reason token.
const backupSeparator = "__"

// ProjectBackup describes one retained backup of a project database.
type ProjectBackup struct {
	ID        string
	ProjectID string
	Reason    string
	CreatedAt time.Time
	Size      int64
	Path      string
}

// BackupsDir is the root directory holding per-project database backups.
func (s Service) BackupsDir() string {
	return filepath.Join(s.DataDir, "backups")
}

func (s Service) projectBackupsDir(projectID string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	if format.IsProjectExtension(projectID) || format.IsSetExtension(projectID) {
		return "", fmt.Errorf("project id must reference an active project, not a database path")
	}
	return filepath.Join(s.BackupsDir(), projectID), nil
}

// backupProjectDatabase snapshots the active project database before a
// destructive operation. It copies the encrypted database bytes, so it requires
// no password and produces a backup encrypted with the same password as the
// live database. It is a no-op when the active database does not exist.
func (s Service) backupProjectDatabase(projectID, reason string) (string, error) {
	source, err := s.ActiveProjectDatabasePath(projectID)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("stat project database: %w", err)
	}

	dir, err := s.projectBackupsDir(projectID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create backups directory: %w", err)
	}

	base := time.Now().UTC().Format(backupTimeLayout) + backupSeparator + reason
	id := base
	target := filepath.Join(dir, id+format.ProjectExtension)
	for n := 2; ; n++ {
		if _, statErr := os.Stat(target); os.IsNotExist(statErr) {
			break
		} else if statErr != nil {
			return "", fmt.Errorf("check backup path: %w", statErr)
		}
		id = fmt.Sprintf("%s%s%d", base, backupSeparator, n)
		target = filepath.Join(dir, id+format.ProjectExtension)
	}

	if err := CopyFile(source, target); err != nil {
		return "", fmt.Errorf("write database backup: %w", err)
	}
	if err := s.pruneProjectBackups(projectID); err != nil {
		return "", err
	}
	return target, nil
}

// ListProjectBackups returns the retained backups for a project, newest first.
func (s Service) ListProjectBackups(projectID string) ([]ProjectBackup, error) {
	dir, err := s.projectBackupsDir(projectID)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ProjectBackup{}, nil
		}
		return nil, fmt.Errorf("read backups directory: %w", err)
	}

	backups := make([]ProjectBackup, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !format.IsProjectExtension(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat backup: %w", err)
		}
		id := strings.TrimSuffix(entry.Name(), format.ProjectExtension)
		createdAt, reason := parseBackupID(id)
		if createdAt.IsZero() {
			createdAt = info.ModTime()
		}
		backups = append(backups, ProjectBackup{
			ID:        id,
			ProjectID: projectID,
			Reason:    reason,
			CreatedAt: createdAt,
			Size:      info.Size(),
			Path:      filepath.Join(dir, entry.Name()),
		})
	}
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ID > backups[j].ID
	})
	return backups, nil
}

// RestoreProjectBackup replaces the active project database with a retained
// backup. The current database, when present, is backed up first, so a restore
// is itself recoverable. Replacing an existing active database requires force.
func (s Service) RestoreProjectBackup(projectID, backupID string, force bool) (string, error) {
	dir, err := s.projectBackupsDir(projectID)
	if err != nil {
		return "", err
	}
	if backupID == "" {
		return "", fmt.Errorf("backup id is required")
	}
	if strings.ContainsAny(backupID, `/\`) || strings.Contains(backupID, "..") {
		return "", fmt.Errorf("invalid backup id")
	}
	backupPath := filepath.Join(dir, backupID+format.ProjectExtension)
	if _, err := os.Stat(backupPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("backup %q not found", backupID)
		}
		return "", fmt.Errorf("stat backup: %w", err)
	}

	target, err := s.ActiveProjectDatabasePath(projectID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", fmt.Errorf("create projects directory: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		if !force {
			return "", fmt.Errorf("active project database exists; overwrite must be confirmed")
		}
		if _, err := s.backupProjectDatabase(projectID, BackupReasonRestore); err != nil {
			return "", err
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat active database: %w", err)
	}

	temp := target + ".restore.tmp"
	_ = os.Remove(temp)
	if err := CopyFile(backupPath, temp); err != nil {
		return "", fmt.Errorf("stage restored database: %w", err)
	}
	if err := os.Rename(temp, target); err != nil {
		_ = os.Remove(temp)
		return "", fmt.Errorf("commit restored database: %w", err)
	}
	return target, nil
}

func (s Service) pruneProjectBackups(projectID string) error {
	keep := s.resolveBackupRetention()
	backups, err := s.ListProjectBackups(projectID)
	if err != nil {
		return err
	}
	if len(backups) <= keep {
		return nil
	}
	for _, backup := range backups[keep:] {
		if err := os.Remove(backup.Path); err != nil {
			return fmt.Errorf("prune database backup: %w", err)
		}
	}
	return nil
}

func (s Service) resolveBackupRetention() int {
	settings, err := s.ReadSettings()
	if err != nil {
		return DefaultBackupRetention
	}
	if settings.BackupRetention > 0 {
		return settings.BackupRetention
	}
	return DefaultBackupRetention
}

func parseBackupID(id string) (time.Time, string) {
	index := strings.Index(id, backupSeparator)
	if index < 0 {
		return time.Time{}, ""
	}
	stamp := id[:index]
	reason := id[index+len(backupSeparator):]
	if cut := strings.Index(reason, backupSeparator); cut >= 0 {
		reason = reason[:cut]
	}
	parsed, err := time.Parse(backupTimeLayout, stamp)
	if err != nil {
		return time.Time{}, reason
	}
	return parsed, reason
}
