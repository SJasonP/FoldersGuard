package app

import (
	"context"
	"fmt"
	"os"

	"foldersguard/internal/project"
)

func (s Service) DecryptProject(ctx context.Context, input DecryptProjectInput) (DecryptProjectResult, error) {
	if err := ValidateExistingDirectory(input.EncryptedRoot, "content"); err != nil {
		return DecryptProjectResult{}, err
	}
	if err := ValidateOutputOutsideSource(input.EncryptedRoot, input.OutputRoot); err != nil {
		return DecryptProjectResult{}, err
	}
	if err := PrepareDirectoryOutput(input.OutputRoot, input.Force, "output"); err != nil {
		return DecryptProjectResult{}, err
	}

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
		return DecryptProjectResult{}, err
	}

	return DecryptProjectResult{
		ProjectID:             plan.Project.ID.String(),
		OutputRoot:            input.OutputRoot,
		DecryptedFiles:        report.DecryptedFiles,
		RestoredFolders:       report.RestoredFolders,
		SkippedFolders:        report.SkippedFolders,
		DeletedEncryptedFiles: deletedEncryptedFiles,
		FailedEncryptedFiles:  0,
	}, nil
}
