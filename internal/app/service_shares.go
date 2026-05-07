package app

import (
	"context"
	"fmt"
	"os"

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
	plan, meta, _, err := s.ReadShareDatabase(ctx, input)
	if err != nil {
		return VerifyResult{}, err
	}
	if meta["database_type"] != "share" {
		return VerifyResult{}, fmt.Errorf("database type = %q, want share", meta["database_type"])
	}
	report, err := (project.Verifier{EncryptedRoot: encryptedRoot}).VerifyContent(ctx, plan)
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
	if err := PrepareDirectoryOutput(input.OutputRoot, input.Force, "output"); err != nil {
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

	report, err := (project.Restorer{
		EncryptedRoot: input.EncryptedRoot,
		OutputRoot:    input.OutputRoot,
		AfterFile:     afterFile,
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
