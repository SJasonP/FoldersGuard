package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"foldersguard/internal/db"
	"foldersguard/internal/progress"
	"foldersguard/internal/storage"
)

func (s Service) ApplyProjectChanges(ctx context.Context, input ApplyProjectChangesInput) (ApplyProjectChangesResult, error) {
	if len(input.RenameChanges) == 0 && len(input.MoveChanges) == 0 && len(input.RemoveChanges) == 0 && len(input.AddChanges) == 0 && len(input.CreateFolderChanges) == 0 {
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

	tracker := progress.FromContext(ctx)
	tracker.SetPhases(progress.PhasePreparing, progress.PhaseEncrypting, progress.PhaseFinalizing)
	tracker.StartPhase(progress.PhasePreparing, false)

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

	contentConnected := input.EncryptedRoot != ""
	if contentConnected {
		if err := ValidateExistingDirectory(input.EncryptedRoot, "encrypted content"); err != nil {
			return ApplyProjectChangesResult{}, err
		}
	}

	contentOperations := make([]ProjectContentOperation, 0, len(input.MoveChanges)+len(input.RemoveChanges)+len(input.AddChanges)+len(input.CreateFolderChanges))
	appliedContentChanges := make([]ProjectContentOperation, 0, len(input.MoveChanges)+len(input.RemoveChanges)+len(input.AddChanges)+len(input.CreateFolderChanges))

	moveChanges := sortedMoveChanges(input.MoveChanges)
	seenMoves := make(map[string]struct{}, len(moveChanges))
	for _, change := range moveChanges {
		if change.ItemPath == "" {
			return ApplyProjectChangesResult{}, fmt.Errorf("move item path is required")
		}
		if change.TargetFolderPath == "" {
			return ApplyProjectChangesResult{}, fmt.Errorf("move target folder path is required")
		}
		moveKey := change.ItemPath + "\x00" + change.TargetFolderPath
		if _, ok := seenMoves[moveKey]; ok {
			return ApplyProjectChangesResult{}, fmt.Errorf("duplicate move for %q", change.ItemPath)
		}
		seenMoves[moveKey] = struct{}{}

		if contentConnected {
			_, operations, err := store.PlanMove(ctx, change.ItemPath, change.TargetFolderPath)
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			var result storage.MoveResult
			applied, err := ApplyStorageContentOperationsWithCommit(operations, ContentOperationApplyOptions{
				ContentRoot: input.EncryptedRoot,
			}, func() error {
				committed, err := store.MoveItem(ctx, change.ItemPath, change.TargetFolderPath, time.Now())
				if err != nil {
					return err
				}
				result = committed
				return nil
			})
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			contentOperations = append(contentOperations, projectContentOperations(result.Operations)...)
			appliedContentChanges = append(appliedContentChanges, appliedProjectContentOperations(applied)...)
		} else {
			result, err := store.MoveItem(ctx, change.ItemPath, change.TargetFolderPath, time.Now())
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			contentOperations = append(contentOperations, projectContentOperations(result.Operations)...)
		}
	}

	removeChanges := sortedRemoveChanges(input.RemoveChanges)
	seenRemoves := make(map[string]struct{}, len(removeChanges))
	for _, change := range removeChanges {
		if change.ItemPath == "" {
			return ApplyProjectChangesResult{}, fmt.Errorf("remove item path is required")
		}
		if _, ok := seenRemoves[change.ItemPath]; ok {
			return ApplyProjectChangesResult{}, fmt.Errorf("duplicate remove for %q", change.ItemPath)
		}
		seenRemoves[change.ItemPath] = struct{}{}

		if contentConnected {
			_, operations, err := store.PlanRemove(ctx, change.ItemPath)
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			var result storage.RemoveResult
			applied, err := ApplyStorageContentOperationsWithCommit(operations, ContentOperationApplyOptions{
				ContentRoot: input.EncryptedRoot,
			}, func() error {
				committed, err := store.RemoveItem(ctx, change.ItemPath, time.Now())
				if err != nil {
					return err
				}
				result = committed
				return nil
			})
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			contentOperations = append(contentOperations, projectContentOperations(result.Operations)...)
			appliedContentChanges = append(appliedContentChanges, appliedProjectContentOperations(applied)...)
		} else {
			result, err := store.RemoveItem(ctx, change.ItemPath, time.Now())
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			contentOperations = append(contentOperations, projectContentOperations(result.Operations)...)
		}
	}

	stagedContentPath := ""
	stagedContentName := ""
	stagedContentOnDesktop := false
	if len(input.AddChanges) > 0 || len(input.CreateFolderChanges) > 0 {
		staging, err := s.prepareProjectChangeStaging(input.ProjectID)
		if err != nil {
			return ApplyProjectChangesResult{}, err
		}
		stagedContentPath = staging.Path
		stagedContentName = staging.Name
		stagedContentOnDesktop = staging.OnDesktop
	}

	addResult, err := s.applyProjectAddChanges(ctx, store, input, stagedContentPath, contentConnected, tracker)
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	if len(addResult.ContentOperations) > 0 {
		contentOperations = append(contentOperations, addResult.ContentOperations...)
		appliedContentChanges = append(appliedContentChanges, addResult.AppliedContentChanges...)
	}

	createFolderResult, err := s.applyProjectCreateFolderChanges(ctx, store, input, stagedContentPath, contentConnected)
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	createFolderChanges := sortedCreateFolderChanges(input.CreateFolderChanges)
	if len(createFolderResult.ContentOperations) > 0 {
		contentOperations = append(contentOperations, createFolderResult.ContentOperations...)
		appliedContentChanges = append(appliedContentChanges, createFolderResult.AppliedContentChanges...)
	}

	tracker.StartPhase(progress.PhaseFinalizing, false)

	resultStagedContentPath := stagedContentPath
	if contentConnected && stagedContentPath != "" {
		if err := os.RemoveAll(stagedContentPath); err != nil {
			return ApplyProjectChangesResult{}, fmt.Errorf("remove uploaded staging content: %w", err)
		}
		resultStagedContentPath = ""
		stagedContentName = ""
		stagedContentOnDesktop = false
	}
	manualContentGuide := !contentConnected && len(contentOperations) > 0

	state, err := s.OpenProjectBrowser(ctx, OpenProjectBrowserInput{
		ProjectID:     input.ProjectID,
		Password:      input.Password,
		EncryptedRoot: input.EncryptedRoot,
	})
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	return ApplyProjectChangesResult{
		ProjectID:              state.ProjectID,
		AppliedRenames:         len(changes),
		AppliedMoves:           len(moveChanges),
		AppliedRemoves:         len(removeChanges),
		AppliedAdds:            len(input.AddChanges),
		AppliedCreatedFolders:  len(createFolderChanges),
		ManualContentGuide:     manualContentGuide,
		StagedContentPath:      resultStagedContentPath,
		StagedContentName:      stagedContentName,
		StagedContentOnDesktop: stagedContentOnDesktop,
		ContentOperations:      contentOperations,
		AppliedContentChanges:  appliedContentChanges,
		BrowserState:           state,
	}, nil
}

func sortedRenameChanges(changes []ProjectRenameChange) []ProjectRenameChange {
	sorted := append([]ProjectRenameChange(nil), changes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return pathDepthForApply(sorted[i].ItemPath) > pathDepthForApply(sorted[j].ItemPath)
	})
	return sorted
}

func sortedMoveChanges(changes []ProjectMoveChange) []ProjectMoveChange {
	sorted := append([]ProjectMoveChange(nil), changes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return pathDepthForApply(sorted[i].ItemPath) > pathDepthForApply(sorted[j].ItemPath)
	})
	return sorted
}

func sortedRemoveChanges(changes []ProjectRemoveChange) []ProjectRemoveChange {
	sorted := append([]ProjectRemoveChange(nil), changes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return pathDepthForApply(sorted[i].ItemPath) > pathDepthForApply(sorted[j].ItemPath)
	})
	return sorted
}

func sortedCreateFolderChanges(changes []ProjectCreateFolderChange) []ProjectCreateFolderChange {
	sorted := append([]ProjectCreateFolderChange(nil), changes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return pathDepthForApply(sorted[i].TargetFolderPath) < pathDepthForApply(sorted[j].TargetFolderPath)
	})
	return sorted
}

func pathDepthForApply(path string) int {
	if path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func projectContentOperations(operations []storage.ContentOperation) []ProjectContentOperation {
	converted := make([]ProjectContentOperation, 0, len(operations))
	for _, operation := range operations {
		converted = append(converted, ProjectContentOperation{
			Type:       operation.Type,
			SourcePath: operation.SourcePath,
			TargetPath: operation.TargetPath,
		})
	}
	return converted
}

func appliedProjectContentOperations(operations []AppliedContentOperation) []ProjectContentOperation {
	converted := make([]ProjectContentOperation, 0, len(operations))
	for _, operation := range operations {
		converted = append(converted, ProjectContentOperation{
			Type:       operation.Type,
			SourcePath: operation.SourcePath,
			TargetPath: operation.TargetPath,
		})
	}
	return converted
}
