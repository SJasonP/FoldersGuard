package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"foldersguard/internal/db"
	"foldersguard/internal/storage"
)

func (s Service) ApplyProjectChanges(ctx context.Context, input ApplyProjectChangesInput) (ApplyProjectChangesResult, error) {
	if len(input.RenameChanges) == 0 {
		state, err := s.OpenProjectBrowser(ctx, OpenProjectBrowserInput{
			ProjectID:     input.ProjectID,
			Password:      input.Password,
			EncryptedRoot: input.EncryptedRoot,
		})
		if err != nil {
			return ApplyProjectChangesResult{}, err
		}
		return ApplyProjectChangesResult{
			ProjectID:    state.ProjectID,
			BrowserState: state,
		}, nil
	}

	databasePath, err := s.ActiveProjectDatabasePath(input.ProjectID)
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	})
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	changes := sortedRenameChanges(input.RenameChanges)
	seen := make(map[string]struct{}, len(changes))
	for _, change := range changes {
		if change.ItemPath == "" {
			return ApplyProjectChangesResult{}, fmt.Errorf("rename item path is required")
		}
		if _, ok := seen[change.ItemPath]; ok {
			return ApplyProjectChangesResult{}, fmt.Errorf("duplicate rename for %q", change.ItemPath)
		}
		seen[change.ItemPath] = struct{}{}
		if _, err := store.RenameItem(ctx, change.ItemPath, change.NewName, time.Now()); err != nil {
			return ApplyProjectChangesResult{}, err
		}
	}

	state, err := s.OpenProjectBrowser(ctx, OpenProjectBrowserInput{
		ProjectID:     input.ProjectID,
		Password:      input.Password,
		EncryptedRoot: input.EncryptedRoot,
	})
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	return ApplyProjectChangesResult{
		ProjectID:      state.ProjectID,
		AppliedRenames: len(changes),
		BrowserState:   state,
	}, nil
}

func sortedRenameChanges(changes []ProjectRenameChange) []ProjectRenameChange {
	sorted := append([]ProjectRenameChange(nil), changes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return pathDepthForApply(sorted[i].ItemPath) > pathDepthForApply(sorted[j].ItemPath)
	})
	return sorted
}

func pathDepthForApply(path string) int {
	if path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}
