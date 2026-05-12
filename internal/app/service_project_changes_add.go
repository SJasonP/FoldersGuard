package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

type projectAddApplyResult struct {
	ContentOperations     []ProjectContentOperation
	AppliedContentChanges []ProjectContentOperation
}

func (s Service) applyProjectAddChanges(ctx context.Context, store *storage.Store, input ApplyProjectChangesInput, stagedContentPath string, contentConnected bool) (projectAddApplyResult, error) {
	if len(input.AddChanges) == 0 {
		return projectAddApplyResult{}, nil
	}
	result := projectAddApplyResult{}

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
	if contentConnected {
		var committed storage.AddResult
		applied, err := ApplyStorageContentOperationsWithCommit(operations, ContentOperationApplyOptions{
			ContentRoot: encryptedRoot,
			StagingRoot: stagedContentPath,
		}, func() error {
			result, err := store.CommitAdd(ctx, change.TargetFolderPath, addition, operations, time.Now())
			if err != nil {
				return err
			}
			committed = result
			return nil
		})
		if err != nil {
			return projectAddApplyResult{}, err
		}
		return projectAddApplyResult{
			ContentOperations:     projectContentOperations(committed.Operations),
			AppliedContentChanges: appliedProjectContentOperations(applied),
		}, nil
	}
	committed, err := store.CommitAdd(ctx, change.TargetFolderPath, addition, operations, time.Now())
	if err != nil {
		return projectAddApplyResult{}, err
	}
	return projectAddApplyResult{ContentOperations: projectContentOperations(committed.Operations)}, nil
}

type projectChangeStaging struct {
	Path      string
	Name      string
	OnDesktop bool
}

func (s Service) prepareProjectChangeStaging(projectID string) (projectChangeStaging, error) {
	stagingRoot := s.StagedContentDir()
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return projectChangeStaging{}, fmt.Errorf("create staged content directory: %w", err)
	}

	projectName := projectID
	names, err := s.readProjectNames()
	if err != nil {
		return projectChangeStaging{}, err
	}
	projectName = s.localProjectName(projectID, names)

	baseName := stagedContentDirectoryName(projectName, time.Now())
	name := baseName
	path := filepath.Join(stagingRoot, name)
	for suffix := 2; ; suffix++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		if err != nil {
			return projectChangeStaging{}, fmt.Errorf("check staged content directory: %w", err)
		}
		name = fmt.Sprintf("%s %d", baseName, suffix)
		path = filepath.Join(stagingRoot, name)
	}
	if err := PrepareDirectoryOutput(path, false, "staged content"); err != nil {
		return projectChangeStaging{}, err
	}
	return projectChangeStaging{
		Path:      path,
		Name:      name,
		OnDesktop: userDesktopDir() != "" && filepath.Clean(stagingRoot) == filepath.Clean(userDesktopDir()),
	}, nil
}

func stagedContentDirectoryName(projectName string, createdAt time.Time) string {
	name := sanitizeStagedContentName(projectName)
	if name == "" {
		name = "FoldersGuard"
	}
	return fmt.Sprintf("%s %s", name, createdAt.Format("2006-01-02 15.04"))
}

func sanitizeStagedContentName(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer(
		":", "-",
		"/", "-",
		"\\", "-",
		"<", "-",
		">", "-",
		"\"", "-",
		"|", "-",
		"?", "-",
		"*", "-",
		"\x00", "",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)
	name = replacer.Replace(name)
	name = strings.Join(strings.Fields(name), " ")
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return strings.Trim(name, ".- ")
}
