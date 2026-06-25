package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
	"foldersguard/internal/progress"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

type projectAddApplyResult struct {
	ContentOperations     []ProjectContentOperation
	AppliedContentChanges []ProjectContentOperation
}

func (s Service) applyProjectAddChanges(ctx context.Context, store *storage.Store, input ApplyProjectChangesInput, stagedContentPath string, contentConnected bool, tracker *progress.Tracker) (projectAddApplyResult, error) {
	if len(input.AddChanges) == 0 {
		return projectAddApplyResult{}, nil
	}
	result := projectAddApplyResult{}

	noiseMode, err := s.resolveNoiseFileHandling("")
	if err != nil {
		return projectAddApplyResult{}, err
	}
	sourceCleanup, err := s.resolveSourceCleanupMode("")
	if err != nil {
		return projectAddApplyResult{}, err
	}

	// Plan every add first so the combined byte and item total is known before
	// any encryption begins. PrepareAdd later assigns storage names but does not
	// change file sizes or counts, so the totals computed here are accurate.
	type plannedAdd struct {
		change   ProjectAddChange
		addition model.PlannedProject
	}
	planned := make([]plannedAdd, 0, len(input.AddChanges))
	var totalBytes int64
	var totalFiles int

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

		maxPartSize, err := s.resolveMaxPartSize(change.MaxPartSize)
		if err != nil {
			return projectAddApplyResult{}, err
		}
		scan, err := fswalk.ScanPathWithNoiseMode(change.SourcePath, noiseMode)
		if err != nil {
			return projectAddApplyResult{}, err
		}
		addition, err := project.AddPlanner{MaxPartSize: maxPartSize}.Plan(scan)
		if err != nil {
			return projectAddApplyResult{}, err
		}
		planned = append(planned, plannedAdd{change: change, addition: addition})
		for _, file := range addition.Files {
			totalBytes += file.OriginalSize
		}
		totalFiles += len(addition.Files)
	}

	tracker.StartPhase(progress.PhaseEncrypting, true)
	tracker.SetTotalBytes(totalBytes)
	tracker.SetTotalItems(totalFiles)

	for _, pa := range planned {
		operations, err := s.applyOnePlannedAdd(ctx, store, pa.change, pa.addition, stagedContentPath, input.EncryptedRoot, contentConnected, sourceCleanup, noiseMode, tracker)
		if err != nil {
			return projectAddApplyResult{}, err
		}
		result.ContentOperations = append(result.ContentOperations, operations.ContentOperations...)
		result.AppliedContentChanges = append(result.AppliedContentChanges, operations.AppliedContentChanges...)
	}

	return result, nil
}

func (s Service) applyOnePlannedAdd(ctx context.Context, store *storage.Store, change ProjectAddChange, addition model.PlannedProject, stagedContentPath, encryptedRoot string, contentConnected bool, sourceCleanup, noiseMode string, tracker *progress.Tracker) (projectAddApplyResult, error) {
	addition, operations, err := store.PrepareAdd(ctx, change.TargetFolderPath, addition)
	if err != nil {
		return projectAddApplyResult{}, err
	}
	if err := (project.Executor{
		OutputRoot:         stagedContentPath,
		Progress:           tracker,
		SkipProgressTotals: true,
	}).EncryptContent(ctx, addition); err != nil {
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
		if err := s.cleanupAddedSources(change.SourcePath, addition, sourceCleanup, noiseMode); err != nil {
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
	if err := s.cleanupAddedSources(change.SourcePath, addition, sourceCleanup, noiseMode); err != nil {
		return projectAddApplyResult{}, err
	}
	return projectAddApplyResult{ContentOperations: projectContentOperations(committed.Operations)}, nil
}

// cleanupAddedSources applies the source-cleanup setting after an add is
// committed. When the setting is delete, the encrypted source files are removed,
// and for a directory add the now-empty source directories are pruned. It runs
// only after a successful commit so an interrupted add never deletes a source
// whose encrypted content was rolled back.
func (s Service) cleanupAddedSources(sourcePath string, addition model.PlannedProject, sourceCleanup, noiseMode string) error {
	if sourceCleanup != SourceCleanupDelete {
		return nil
	}
	for _, file := range addition.Files {
		if file.SourcePath == "" {
			continue
		}
		if err := os.Remove(file.SourcePath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("delete source file: %w", err)
		}
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat add source: %w", err)
	}
	if info.IsDir() {
		if _, err := removeEmptyFoldersUnderRoot(sourcePath, noiseMode); err != nil {
			return err
		}
	}
	return nil
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
