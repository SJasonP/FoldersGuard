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
	if input.Resume {
		// Resuming keeps the existing partial output and skips already-restored
		// files, so the non-empty output must not be rejected or wiped.
		if err := os.MkdirAll(input.OutputRoot, 0o755); err != nil {
			return DecryptProjectResult{}, fmt.Errorf("create output folder: %w", err)
		}
	} else if err := PrepareDirectoryOutputWithNoiseMode(input.OutputRoot, input.Force, "output", noiseMode); err != nil {
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
	deletedEncryptedFiles := 0
	afterFile := func(restored project.RestoredFile) error {
		if sourceCleanup != SourceCleanupDelete {
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
		EncryptedRoot:   input.EncryptedRoot,
		OutputRoot:      input.OutputRoot,
		NoiseMode:       noiseMode,
		AfterFile:       afterFile,
		Progress:        tracker,
		Resume:          input.Resume,
		ContinueOnError: continueOnError,
		OnFileError:     onFileError,
	}).RestoreContentReport(ctx, plan)
	if err != nil {
		return DecryptProjectResult{}, err
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
