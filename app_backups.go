package main

import (
	"time"
)

// ListProjectBackups returns the retained database backups for a project,
// newest first. Listing reads only the backups directory and needs no password.
func (a *App) ListProjectBackups(projectID string) ([]ProjectBackupInfo, error) {
	backups, err := a.service.ListProjectBackups(projectID)
	if err != nil {
		return nil, frontendError(err)
	}
	result := make([]ProjectBackupInfo, 0, len(backups))
	for _, backup := range backups {
		createdAt := ""
		if !backup.CreatedAt.IsZero() {
			createdAt = backup.CreatedAt.Format(time.RFC3339)
		}
		result = append(result, ProjectBackupInfo{
			ID:        backup.ID,
			ProjectID: backup.ProjectID,
			Reason:    backup.Reason,
			CreatedAt: createdAt,
			Size:      backup.Size,
		})
	}
	return result, nil
}

// RestoreProjectBackup replaces the active project database with a retained
// backup. The current database is backed up first, and overwriting it requires
// force, which the WebUI sets after explicit confirmation.
func (a *App) RestoreProjectBackup(request RestoreProjectBackupRequest) (RestoreProjectBackupResult, error) {
	if _, err := a.service.RestoreProjectBackup(request.ProjectID, request.BackupID, request.Force); err != nil {
		return RestoreProjectBackupResult{}, frontendError(err)
	}
	return RestoreProjectBackupResult{ProjectID: request.ProjectID}, nil
}
