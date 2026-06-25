package main

import (
	"time"

	"foldersguard/internal/app"
)

func (a *App) ListLocalProjects() ([]LocalProjectSummary, error) {
	projects, err := a.service.ListActiveProjects()
	if err != nil {
		return nil, frontendError(err)
	}

	result := make([]LocalProjectSummary, 0, len(projects))
	for _, project := range projects {
		modifiedAt := ""
		if !project.ModifiedAt.IsZero() {
			modifiedAt = project.ModifiedAt.Format(time.RFC3339)
		}
		result = append(result, LocalProjectSummary{
			ProjectID:          project.ProjectID,
			ProjectName:        project.ProjectName,
			FileName:           project.FileName,
			ModifiedAt:         modifiedAt,
			AvailabilityStatus: project.Availability,
		})
	}
	return result, nil
}

func (a *App) SaveLocalProjectName(request SaveLocalProjectNameRequest) (SaveLocalProjectNameResult, error) {
	result, err := a.service.SaveLocalProjectName(app.SaveLocalProjectNameInput{
		ProjectID:   request.ProjectID,
		ProjectName: request.ProjectName,
	})
	if err != nil {
		return SaveLocalProjectNameResult{}, frontendError(err)
	}
	return SaveLocalProjectNameResult{
		ProjectID:   result.ProjectID,
		ProjectName: result.ProjectName,
	}, nil
}

func (a *App) InspectProject(request InspectProjectRequest) (InspectProjectResult, error) {
	result, err := a.service.Inspect(a.ctx, app.DatabaseOpen{
		ProjectRef: request.ProjectID,
		Password:   request.Password,
	})
	if err != nil {
		return InspectProjectResult{}, frontendError(err)
	}
	return InspectProjectResult{
		ProjectID:      result.ProjectID,
		DatabaseType:   result.DatabaseType,
		ProjectName:    result.ProjectName,
		RootFolderID:   result.RootFolderID,
		RootName:       result.RootName,
		FormatVersion:  result.FormatVersion,
		CreatedAt:      result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      result.UpdatedAt.Format(time.RFC3339),
		Items:          result.Items,
		Folders:        result.Folders,
		Files:          result.Files,
		Parts:          result.Parts,
		StorageObjects: result.StorageObjects,
	}, nil
}

func (a *App) VerifyProject(request VerifyProjectRequest) (VerifyProjectResult, error) {
	ctx, finish := a.beginOperation("verify")
	result, err := a.service.Verify(ctx, app.DatabaseOpen{
		ProjectRef: request.ProjectID,
		Password:   request.Password,
	}, request.EncryptedPath)
	finish(err)
	if err != nil {
		return VerifyProjectResult{}, frontendError(err)
	}
	return VerifyProjectResult{
		ProjectID:       result.ProjectID,
		CheckedObjects:  result.CheckedObjects,
		MissingObjects:  result.MissingObjects,
		TamperedObjects: result.TamperedObjects,
		ExtraObjects:    result.ExtraObjects,
		MissingPaths:    result.MissingPaths,
		TamperedPaths:   result.TamperedPaths,
		ExtraPaths:      result.ExtraPaths,
		Status:          result.Status,
	}, nil
}

func (a *App) DecryptProject(request DecryptProjectRequest) (DecryptProjectResult, error) {
	ctx, finish := a.beginOperation("decrypt")
	result, err := a.service.DecryptProject(ctx, app.DecryptProjectInput{
		ProjectID:     request.ProjectID,
		Password:      request.Password,
		EncryptedRoot: request.EncryptedPath,
		OutputRoot:    request.OutputPath,
		Force:         request.Force,
		SourceCleanup: request.SourceCleanup,
		Resume:        request.Resume,
	})
	finish(err)
	if err != nil {
		return DecryptProjectResult{}, frontendError(err)
	}
	return DecryptProjectResult{
		ProjectID:             result.ProjectID,
		OutputPath:            result.OutputRoot,
		DecryptedFiles:        result.DecryptedFiles,
		RestoredFolders:       result.RestoredFolders,
		SkippedFolders:        result.SkippedFolders,
		DeletedEncryptedFiles: result.DeletedEncryptedFiles,
		FailedEncryptedFiles:  result.FailedEncryptedFiles,
	}, nil
}

func (a *App) LoadShare(request LoadShareRequest) (ShareSummary, error) {
	result, err := a.service.InspectShare(a.ctx, app.ShareOpen{
		DatabasePath: request.DatabasePath,
		Password:     request.Password,
	})
	if err != nil {
		return ShareSummary{}, frontendError(err)
	}
	return ShareSummary{
		ShareID:           result.ShareID,
		DatabaseType:      result.DatabaseType,
		FormatVersion:     result.FormatVersion,
		TopLevelItems:     result.TopLevelItems,
		Files:             result.Files,
		Folders:           result.Folders,
		Parts:             result.Parts,
		StorageObjects:    result.StorageObjects,
		PasswordProtected: result.PasswordProtected,
	}, nil
}

func (a *App) VerifyShare(request VerifyShareRequest) (VerifyProjectResult, error) {
	ctx, finish := a.beginOperation("verify")
	result, err := a.service.VerifyShare(ctx, app.ShareOpen{
		DatabasePath: request.DatabasePath,
		Password:     request.Password,
	}, request.EncryptedPath)
	finish(err)
	if err != nil {
		return VerifyProjectResult{}, frontendError(err)
	}
	return VerifyProjectResult{
		ProjectID:       result.ProjectID,
		CheckedObjects:  result.CheckedObjects,
		MissingObjects:  result.MissingObjects,
		TamperedObjects: result.TamperedObjects,
		ExtraObjects:    result.ExtraObjects,
		MissingPaths:    result.MissingPaths,
		TamperedPaths:   result.TamperedPaths,
		ExtraPaths:      result.ExtraPaths,
		Status:          result.Status,
	}, nil
}

func (a *App) DecryptShare(request DecryptShareRequest) (DecryptShareResult, error) {
	ctx, finish := a.beginOperation("decrypt")
	result, err := a.service.DecryptShare(ctx, app.DecryptShareInput{
		DatabasePath:  request.DatabasePath,
		Password:      request.Password,
		EncryptedRoot: request.EncryptedPath,
		OutputRoot:    request.OutputPath,
		Force:         request.Force,
		SourceCleanup: request.SourceCleanup,
		Resume:        request.Resume,
	})
	finish(err)
	if err != nil {
		return DecryptShareResult{}, frontendError(err)
	}
	return DecryptShareResult{
		ShareID:               result.ShareID,
		OutputPath:            result.OutputRoot,
		DecryptedFiles:        result.DecryptedFiles,
		RestoredFolders:       result.RestoredFolders,
		SkippedFolders:        result.SkippedFolders,
		DeletedEncryptedFiles: result.DeletedEncryptedFiles,
		FailedEncryptedFiles:  result.FailedEncryptedFiles,
	}, nil
}

func (a *App) ExportProject(request ExportProjectRequest) (ExportProjectResult, error) {
	ctx, finish := a.beginOperation("export")
	result, err := a.service.ExportProject(ctx, app.ExportProjectInput{
		ProjectID:  request.ProjectID,
		Password:   request.Password,
		OutputPath: request.OutputPath,
		Force:      request.Force,
	})
	finish(err)
	if err != nil {
		return ExportProjectResult{}, frontendError(err)
	}
	return ExportProjectResult{
		ProjectID:  result.ProjectID,
		OutputPath: result.OutputPath,
	}, nil
}

func (a *App) DeleteProject(request DeleteProjectRequest) (DeleteProjectResult, error) {
	result, err := a.service.DeleteProject(a.ctx, app.DeleteProjectInput{
		ProjectID: request.ProjectID,
		Password:  request.Password,
	})
	if err != nil {
		return DeleteProjectResult{}, frontendError(err)
	}
	return DeleteProjectResult{
		ProjectID: result.ProjectID,
	}, nil
}

func (a *App) CreateProject(request CreateProjectRequest) (CreateProjectResult, error) {
	ctx, finish := a.beginOperation("create")
	result, err := a.service.CreateProject(ctx, app.CreateProjectInput{
		SourcePath:     request.SourcePath,
		ContentOutput:  request.ContentOutput,
		Password:       request.Password,
		MaxPartSize:    request.MaxPartSize,
		Force:          request.Force,
		SourceCleanup:  request.SourceCleanup,
		DatabaseExport: request.DatabaseExport,
	})
	finish(err)
	if err != nil {
		return CreateProjectResult{}, frontendError(err)
	}
	return CreateProjectResult{
		ProjectID:               result.ProjectID,
		ProjectName:             result.ProjectName,
		ContentOutput:           result.ContentOutput,
		DatabaseExport:          result.DatabaseExport,
		EncryptedFiles:          result.EncryptedFiles,
		EncryptedFolders:        result.EncryptedFolders,
		EncryptedParts:          result.EncryptedParts,
		DeletedCleartextFiles:   result.DeletedCleartextFiles,
		DeletedCleartextFolders: result.DeletedCleartextFolders,
		FailedFiles:             result.FailedFiles,
	}, nil
}

func (a *App) ImportProject(request ImportProjectRequest) (ImportProjectResult, error) {
	ctx, finish := a.beginOperation("import")
	result, err := a.service.ImportProject(ctx, app.ImportProjectInput{
		InputPath: request.InputPath,
		Password:  request.Password,
		Force:     request.Force,
	})
	finish(err)
	if err != nil {
		return ImportProjectResult{}, frontendError(err)
	}
	return ImportProjectResult{
		ProjectID: result.ProjectID,
	}, nil
}
