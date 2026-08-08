package app

import (
	"context"
	"fmt"
	"os"

	"foldersguard/internal/model"
	"foldersguard/internal/progress"
	"foldersguard/internal/project"
)

func (s Service) DecryptProject(ctx context.Context, input DecryptProjectInput) (DecryptProjectResult, error) {
	if err := ValidateExistingDirectory(input.EncryptedRoot, "content"); err != nil {
		return DecryptProjectResult{}, err
	}
	if err := ValidateOutputOutsideSource(input.EncryptedRoot, input.OutputRoot); err != nil {
		return DecryptProjectResult{}, err
	}
	noiseMode, err := s.resolveNoiseFileHandling("")
	if err != nil {
		return DecryptProjectResult{}, err
	}
	tracker := progress.FromContext(ctx)
	tracker.SetPhases(progress.PhasePreparing, progress.PhaseDecrypting)
	tracker.StartPhase(progress.PhasePreparing, false)

	plan, meta, err := s.ReadDatabase(ctx, DatabaseOpen{
		ProjectRef: input.ProjectID,
		Password:   input.Password,
	})
	if err != nil {
		return DecryptProjectResult{}, err
	}
	if meta["database_type"] != "project" {
		return DecryptProjectResult{}, fmt.Errorf("database type = %q, want project", meta["database_type"])
	}

	sourceCleanup, err := s.resolveSourceCleanupMode(input.SourceCleanup)
	if err != nil {
		return DecryptProjectResult{}, err
	}
	failureHandling, err := s.resolveFailureHandling(input.FailureHandling)
	if err != nil {
		return DecryptProjectResult{}, err
	}
	progressiveCleanup := sourceCleanup == SourceCleanupAfterFile || sourceCleanup == SourceCleanupAfterPart
	partCleanup := sourceCleanup == SourceCleanupAfterPart
	restorer := project.Restorer{EncryptedRoot: input.EncryptedRoot, OutputRoot: input.OutputRoot, NoiseMode: noiseMode, Resume: input.Resume, IncrementalCleanup: partCleanup}
	outputSizes, sourceSizes, err := restorer.RestoreSpaceSizes(ctx, plan)
	if err != nil {
		return DecryptProjectResult{}, err
	}
	if err := ensureOperationSpace(input.EncryptedRoot, input.OutputRoot, outputSizes, sourceSizes, progressiveCleanup); err != nil {
		return DecryptProjectResult{}, err
	}
	if input.Resume {
		// Resuming keeps the existing partial output and skips already-restored
		// files, so the non-empty output must not be rejected or wiped.
		if err := os.MkdirAll(input.OutputRoot, 0o755); err != nil {
			return DecryptProjectResult{}, fmt.Errorf("create output folder: %w", err)
		}
	} else if err := PrepareDirectoryOutputWithNoiseMode(input.OutputRoot, input.Force, "output", noiseMode); err != nil {
		return DecryptProjectResult{}, err
	}
	deletedEncryptedFiles := 0
	var successfullyRestored []project.RestoredFile
	afterFile := func(restored project.RestoredFile) error {
		successfullyRestored = append(successfullyRestored, restored)
		if !progressiveCleanup {
			return nil
		}
		if partCleanup && restored.File.StorageKind == model.StorageKindSplit {
			return nil
		}
		for _, path := range restored.EncryptedAbsolutePaths {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("delete encrypted file: %w", err)
			}
			deletedEncryptedFiles++
		}
		return nil
	}
	afterPart := func(restored project.RestoredPart) error {
		if !partCleanup {
			return nil
		}
		if err := os.Remove(restored.EncryptedAbsPath); err != nil {
			return fmt.Errorf("delete encrypted part: %w", err)
		}
		deletedEncryptedFiles++
		return nil
	}

	continueOnError := failureHandling == FailureHandlingContinue
	var failures []FailedItem
	onFileError := func(file model.File, ferr error) {
		failures = append(failures, FailedItem{
			FileID: file.ID.String(),
			Reason: ferr.Error(),
		})
	}

	tracker.StartPhase(progress.PhaseDecrypting, true)
	report, err := (project.Restorer{
		EncryptedRoot:      input.EncryptedRoot,
		OutputRoot:         input.OutputRoot,
		NoiseMode:          noiseMode,
		AfterFile:          afterFile,
		AfterPart:          afterPart,
		IncrementalCleanup: partCleanup,
		Progress:           tracker,
		Resume:             input.Resume,
		ContinueOnError:    continueOnError,
		OnFileError:        onFileError,
	}).RestoreContentReport(ctx, plan)
	if err != nil {
		return DecryptProjectResult{}, err
	}
	if sourceCleanup == SourceCleanupAfterOperation && len(failures) == 0 {
		for _, restored := range successfullyRestored {
			for _, path := range restored.EncryptedAbsolutePaths {
				if err := os.Remove(path); err != nil {
					return DecryptProjectResult{}, fmt.Errorf("delete encrypted file: %w", err)
				}
				deletedEncryptedFiles++
			}
		}
	}

	return DecryptProjectResult{
		ProjectID:             plan.Project.ID.String(),
		OutputRoot:            input.OutputRoot,
		DecryptedFiles:        report.DecryptedFiles,
		RestoredFolders:       report.RestoredFolders,
		SkippedFolders:        report.SkippedFolders,
		DeletedEncryptedFiles: deletedEncryptedFiles,
		FailedEncryptedFiles:  len(failures),
		Failures:              failures,
	}, nil
}
