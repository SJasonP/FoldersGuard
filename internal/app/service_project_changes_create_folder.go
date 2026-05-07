package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"foldersguard/internal/content"
	"foldersguard/internal/storage"
)

func (s Service) applyProjectCreateFolderChanges(ctx context.Context, store *storage.Store, input ApplyProjectChangesInput, stagedContentPath string, contentConnected bool) (projectAddApplyResult, error) {
	if len(input.CreateFolderChanges) == 0 {
		return projectAddApplyResult{}, nil
	}

	result := projectAddApplyResult{}
	createFolderChanges := sortedCreateFolderChanges(input.CreateFolderChanges)
	seenCreateFolders := make(map[string]struct{}, len(createFolderChanges))
	for _, change := range createFolderChanges {
		if change.TargetFolderPath == "" {
			return projectAddApplyResult{}, fmt.Errorf("create folder target path is required")
		}
		if change.Name == "" {
			return projectAddApplyResult{}, fmt.Errorf("create folder name is required")
		}
		createKey := change.TargetFolderPath + "\x00" + change.Name
		if _, ok := seenCreateFolders[createKey]; ok {
			return projectAddApplyResult{}, fmt.Errorf("duplicate create folder %q in %q", change.Name, change.TargetFolderPath)
		}
		seenCreateFolders[createKey] = struct{}{}

		created, err := s.applyOneProjectCreateFolder(ctx, store, change, stagedContentPath, input.EncryptedRoot, contentConnected)
		if err != nil {
			return projectAddApplyResult{}, err
		}
		result.ContentOperations = append(result.ContentOperations, created.ContentOperations...)
		result.AppliedContentChanges = append(result.AppliedContentChanges, created.AppliedContentChanges...)
	}
	return result, nil
}

func (s Service) applyOneProjectCreateFolder(ctx context.Context, store *storage.Store, change ProjectCreateFolderChange, stagedContentPath, encryptedRoot string, contentConnected bool) (projectAddApplyResult, error) {
	prepared, err := store.PrepareCreateFolder(ctx, change.TargetFolderPath, change.Name, time.Now())
	if err != nil {
		return projectAddApplyResult{}, err
	}
	stagedFolder, err := content.SafeJoin(stagedContentPath, prepared.Operation.SourcePath)
	if err != nil {
		return projectAddApplyResult{}, fmt.Errorf("resolve staged created folder: %w", err)
	}
	if err := os.MkdirAll(stagedFolder, 0o755); err != nil {
		return projectAddApplyResult{}, fmt.Errorf("create staged folder content: %w", err)
	}
	if contentConnected {
		if err := ValidateStorageContentOperations([]storage.ContentOperation{prepared.Operation}, ContentOperationApplyOptions{
			ContentRoot: encryptedRoot,
			StagingRoot: stagedContentPath,
		}); err != nil {
			return projectAddApplyResult{}, err
		}
	}

	committed, err := store.CommitCreateFolder(ctx, prepared, time.Now())
	if err != nil {
		return projectAddApplyResult{}, err
	}
	result := projectAddApplyResult{ContentOperations: projectContentOperations(committed.Operations)}
	if contentConnected {
		applied, err := ApplyStorageContentOperations(committed.Operations, ContentOperationApplyOptions{
			ContentRoot: encryptedRoot,
			StagingRoot: stagedContentPath,
		})
		if err != nil {
			return projectAddApplyResult{}, err
		}
		result.AppliedContentChanges = appliedProjectContentOperations(applied)
	}
	return result, nil
}
