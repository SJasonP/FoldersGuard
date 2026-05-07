package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

type projectAddApplyResult struct {
	StagedContentPath     string
	ContentOperations     []ProjectContentOperation
	AppliedContentChanges []ProjectContentOperation
}

func (s Service) applyProjectAddChanges(ctx context.Context, store *storage.Store, input ApplyProjectChangesInput, contentConnected bool) (projectAddApplyResult, error) {
	if len(input.AddChanges) == 0 {
		return projectAddApplyResult{}, nil
	}

	stagedContentPath, err := s.prepareProjectChangeStaging(input.ProjectID)
	if err != nil {
		return projectAddApplyResult{}, err
	}
	result := projectAddApplyResult{StagedContentPath: stagedContentPath}

	seenAdds := make(map[string]struct{}, len(input.AddChanges))
	for _, change := range input.AddChanges {
		if change.SourcePath == "" {
			return projectAddApplyResult{}, fmt.Errorf("add source path is required")
		}
		if change.TargetFolderPath == "" {
			return projectAddApplyResult{}, fmt.Errorf("add target folder path is required")
		}
		addKey := change.SourcePath + "\x00" + change.TargetFolderPath
		if _, ok := seenAdds[addKey]; ok {
			return projectAddApplyResult{}, fmt.Errorf("duplicate add for %q", change.SourcePath)
		}
		seenAdds[addKey] = struct{}{}

		operations, err := s.applyOneProjectAdd(ctx, store, change, stagedContentPath, input.EncryptedRoot, contentConnected)
		if err != nil {
			return projectAddApplyResult{}, err
		}
		result.ContentOperations = append(result.ContentOperations, operations.ContentOperations...)
		result.AppliedContentChanges = append(result.AppliedContentChanges, operations.AppliedContentChanges...)
	}

	if contentConnected {
		if err := os.RemoveAll(stagedContentPath); err != nil {
			return projectAddApplyResult{}, fmt.Errorf("remove uploaded staging content: %w", err)
		}
		result.StagedContentPath = ""
	}
	return result, nil
}

func (s Service) applyOneProjectAdd(ctx context.Context, store *storage.Store, change ProjectAddChange, stagedContentPath, encryptedRoot string, contentConnected bool) (projectAddApplyResult, error) {
	maxPartSize, err := s.resolveMaxPartSize(change.MaxPartSize)
	if err != nil {
		return projectAddApplyResult{}, err
	}
	scan, err := fswalk.ScanPath(change.SourcePath)
	if err != nil {
		return projectAddApplyResult{}, err
	}
	addition, err := project.AddPlanner{MaxPartSize: maxPartSize}.Plan(scan)
	if err != nil {
		return projectAddApplyResult{}, err
	}
	addition, operations, err := store.PrepareAdd(ctx, change.TargetFolderPath, addition)
	if err != nil {
		return projectAddApplyResult{}, err
	}
	if err := (project.Executor{OutputRoot: stagedContentPath}).EncryptContent(ctx, addition); err != nil {
		return projectAddApplyResult{}, err
	}
	if contentConnected {
		if err := ValidateStorageContentOperations(operations, ContentOperationApplyOptions{
			ContentRoot: encryptedRoot,
			StagingRoot: stagedContentPath,
		}); err != nil {
			return projectAddApplyResult{}, err
		}
	}
	committed, err := store.CommitAdd(ctx, change.TargetFolderPath, addition, operations, time.Now())
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

func (s Service) prepareProjectChangeStaging(projectID string) (string, error) {
	if err := os.MkdirAll(s.StagedContentDir(), 0o755); err != nil {
		return "", fmt.Errorf("create staged content directory: %w", err)
	}
	path := filepath.Join(s.StagedContentDir(), projectID+"-"+uuid.NewString())
	if err := PrepareDirectoryOutput(path, false, "staged content"); err != nil {
		return "", err
	}
	return path, nil
}
