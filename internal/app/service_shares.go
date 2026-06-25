package app

import (
	"context"
	"fmt"
	"os"

	"foldersguard/internal/progress"
	"foldersguard/internal/project"
)

func (s Service) InspectShare(ctx context.Context, input ShareOpen) (ShareSummary, error) {
	plan, meta, passwordProtected, err := s.ReadShareDatabase(ctx, input)
	if err != nil {
		return ShareSummary{}, err
	}
	if meta["database_type"] != "share" {
		return ShareSummary{}, fmt.Errorf("database type = %q, want share", meta["database_type"])
	}
	return shareSummary(plan, meta, passwordProtected), nil
}

func (s Service) VerifyShare(ctx context.Context, input ShareOpen, encryptedRoot string) (VerifyResult, error) {
	if err := ValidateExistingDirectory(encryptedRoot, "content"); err != nil {
		return VerifyResult{}, err
	}
	noiseMode, err := s.resolveNoiseFileHandling("")
	if err != nil {
		return VerifyResult{}, err
	}
	plan, meta, _, err := s.ReadShareDatabase(ctx, input)
	if err != nil {
		return VerifyResult{}, err
	}
	if meta["database_type"] != "share" {
		return VerifyResult{}, fmt.Errorf("database type = %q, want share", meta["database_type"])
	}
	tracker := progress.FromContext(ctx)
	tracker.SetPhases(progress.PhaseVerifying)
	tracker.StartPhase(progress.PhaseVerifying, true)
	report, err := (project.Verifier{EncryptedRoot: encryptedRoot, NoiseMode: noiseMode, Progress: tracker}).VerifyContent(ctx, plan)
	if err != nil {
		return VerifyResult{}, err
	}
	status := "ok"
	if !report.OK() {
		status = "failed"
	}
	return VerifyResult{
		ProjectID:       plan.Project.ID.String(),
		CheckedObjects:  report.CheckedObjects,
		MissingObjects:  report.MissingObjects,
		TamperedObjects: report.TamperedObjects,
		ExtraObjects:    report.ExtraObjects,
		MissingPaths:    report.MissingPaths,
		TamperedPaths:   report.TamperedPaths,
		ExtraPaths:      report.ExtraPaths,
		Status:          status,
	}, nil
}

func (s Service) DecryptShare(ctx context.Context, input DecryptShareInput) (DecryptShareResult, error) {
	if err := ValidateExistingDirectory(input.EncryptedRoot, "content"); err != nil {
		return DecryptShareResult{}, err
	}
	if err := ValidateOutputOutsideSource(input.EncryptedRoot, input.OutputRoot); err != nil {
		return DecryptShareResult{}, err
	}
	noiseMode, err := s.resolveNoiseFileHandling("")
	if err != nil {
		return DecryptShareResult{}, err
	}
	if input.Resume {
		if err := os.MkdirAll(input.OutputRoot, 0o755); err != nil {
			return DecryptShareResult{}, fmt.Errorf("create output folder: %w", err)
		}
	} else if err := PrepareDirectoryOutputWithNoiseMode(input.OutputRoot, input.Force, "output", noiseMode); err != nil {
		return DecryptShareResult{}, err
	}

	plan, meta, _, err := s.ReadShareDatabase(ctx, ShareOpen{
		DatabasePath: input.DatabasePath,
		Password:     input.Password,
	})
	if err != nil {
		return DecryptShareResult{}, err
	}
	if meta["database_type"] != "share" {
		return DecryptShareResult{}, fmt.Errorf("database type = %q, want share", meta["database_type"])
	}

	sourceCleanup, err := s.resolveSourceCleanupMode(input.SourceCleanup)
	if err != nil {
		return DecryptShareResult{}, err
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

	tracker := progress.FromContext(ctx)
	tracker.SetPhases(progress.PhaseDecrypting)
	tracker.StartPhase(progress.PhaseDecrypting, true)
	report, err := (project.Restorer{
		EncryptedRoot: input.EncryptedRoot,
		OutputRoot:    input.OutputRoot,
		NoiseMode:     noiseMode,
		AfterFile:     afterFile,
		Progress:      tracker,
		Resume:        input.Resume,
	}).RestoreContentReport(ctx, plan)
	if err != nil {
		return DecryptShareResult{}, err
	}

	return DecryptShareResult{
		ShareID:               plan.Project.ID.String(),
		OutputRoot:            input.OutputRoot,
		DecryptedFiles:        report.DecryptedFiles,
		RestoredFolders:       report.RestoredFolders,
		SkippedFolders:        report.SkippedFolders,
		DeletedEncryptedFiles: deletedEncryptedFiles,
		FailedEncryptedFiles:  0,
	}, nil
}
