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
	if len(input.RenameChanges) == 0 && len(input.MoveChanges) == 0 && len(input.RemoveChanges) == 0 && len(input.AddChanges) == 0 {
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

	contentConnected := input.EncryptedRoot != ""
	if contentConnected {
		if err := ValidateExistingDirectory(input.EncryptedRoot, "encrypted content"); err != nil {
			return ApplyProjectChangesResult{}, err
		}
	}

	contentOperations := make([]ProjectContentOperation, 0, len(input.MoveChanges)+len(input.RemoveChanges)+len(input.AddChanges))
	appliedContentChanges := make([]ProjectContentOperation, 0, len(input.MoveChanges)+len(input.RemoveChanges)+len(input.AddChanges))

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
			if _, operations, err := store.PlanMove(ctx, change.ItemPath, change.TargetFolderPath); err != nil {
				return ApplyProjectChangesResult{}, err
			} else if err := ValidateStorageContentOperations(operations, ContentOperationApplyOptions{
				ContentRoot: input.EncryptedRoot,
			}); err != nil {
				return ApplyProjectChangesResult{}, err
			}
		}
		result, err := store.MoveItem(ctx, change.ItemPath, change.TargetFolderPath, time.Now())
		if err != nil {
			return ApplyProjectChangesResult{}, err
		}
		contentOperations = append(contentOperations, projectContentOperations(result.Operations)...)
		if contentConnected {
			applied, err := ApplyStorageContentOperations(result.Operations, ContentOperationApplyOptions{
				ContentRoot: input.EncryptedRoot,
			})
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			appliedContentChanges = append(appliedContentChanges, appliedProjectContentOperations(applied)...)
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
			if _, operations, err := store.PlanRemove(ctx, change.ItemPath); err != nil {
				return ApplyProjectChangesResult{}, err
			} else if err := ValidateStorageContentOperations(operations, ContentOperationApplyOptions{
				ContentRoot: input.EncryptedRoot,
			}); err != nil {
				return ApplyProjectChangesResult{}, err
			}
		}
		result, err := store.RemoveItem(ctx, change.ItemPath, time.Now())
		if err != nil {
			return ApplyProjectChangesResult{}, err
		}
		contentOperations = append(contentOperations, projectContentOperations(result.Operations)...)
		if contentConnected {
			applied, err := ApplyStorageContentOperations(result.Operations, ContentOperationApplyOptions{
				ContentRoot: input.EncryptedRoot,
			})
			if err != nil {
				return ApplyProjectChangesResult{}, err
			}
			appliedContentChanges = append(appliedContentChanges, appliedProjectContentOperations(applied)...)
		}
	}

	addResult, err := s.applyProjectAddChanges(ctx, store, input, contentConnected)
	if err != nil {
		return ApplyProjectChangesResult{}, err
	}
	if len(addResult.ContentOperations) > 0 {
		contentOperations = append(contentOperations, addResult.ContentOperations...)
		appliedContentChanges = append(appliedContentChanges, addResult.AppliedContentChanges...)
	}

	operationGuidePath := ""
	if !contentConnected && len(contentOperations) > 0 {
		settings, err := s.ReadSettings()
		if err != nil {
			return ApplyProjectChangesResult{}, err
		}
		path, err := s.WriteOperationGuide(OperationGuideInput{
			ProjectID:  input.ProjectID,
			Operations: contentOperations,
			CreatedAt:  time.Now(),
			Format:     settings.OperationGuideFormat,
		})
		if err != nil {
			return ApplyProjectChangesResult{}, err
		}
		operationGuidePath = path
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
		ProjectID:             state.ProjectID,
		AppliedRenames:        len(changes),
		AppliedMoves:          len(moveChanges),
		AppliedRemoves:        len(removeChanges),
		AppliedAdds:           len(input.AddChanges),
		OperationGuidePath:    operationGuidePath,
		StagedContentPath:     addResult.StagedContentPath,
		ContentOperations:     contentOperations,
		AppliedContentChanges: appliedContentChanges,
		BrowserState:          state,
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
