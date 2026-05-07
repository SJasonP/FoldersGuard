package main

import (
	"time"

	"foldersguard/internal/app"
)

func (a *App) ListLocalProjects() ([]LocalProjectSummary, error) {
	projects, err := a.service.ListActiveProjects()
	if err != nil {
		return nil, err
	}

	result := make([]LocalProjectSummary, 0, len(projects))
	for _, project := range projects {
		modifiedAt := ""
		if !project.ModifiedAt.IsZero() {
			modifiedAt = project.ModifiedAt.Format(time.RFC3339)
		}
		result = append(result, LocalProjectSummary{
			ProjectID:          project.ProjectID,
			FileName:           project.FileName,
			ModifiedAt:         modifiedAt,
			AvailabilityStatus: project.Availability,
		})
	}
	return result, nil
}

func (a *App) InspectProject(request InspectProjectRequest) (InspectProjectResult, error) {
	result, err := a.service.Inspect(a.ctx, app.DatabaseOpen{
		ProjectRef: request.ProjectID,
		Password:   request.Password,
	})
	if err != nil {
		return InspectProjectResult{}, err
	}
	return InspectProjectResult{
		ProjectID:      result.ProjectID,
		DatabaseType:   result.DatabaseType,
		ProjectName:    result.ProjectName,
		RootFolderID:   result.RootFolderID,
		RootName:       result.RootName,
		FormatVersion:  result.FormatVersion,
		SchemaVersion:  result.SchemaVersion,
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
	result, err := a.service.Verify(a.ctx, app.DatabaseOpen{
		ProjectRef: request.ProjectID,
		Password:   request.Password,
	}, request.EncryptedPath)
	if err != nil {
		return VerifyProjectResult{}, err
	}
	return VerifyProjectResult{
		ProjectID:       result.ProjectID,
		CheckedObjects:  result.CheckedObjects,
		MissingObjects:  result.MissingObjects,
		TamperedObjects: result.TamperedObjects,
		ExtraObjects:    result.ExtraObjects,
		Status:          result.Status,
	}, nil
}

func (a *App) DecryptProject(request DecryptProjectRequest) (DecryptProjectResult, error) {
	result, err := a.service.DecryptProject(a.ctx, app.DecryptProjectInput{
		ProjectID:     request.ProjectID,
		Password:      request.Password,
		EncryptedRoot: request.EncryptedPath,
		OutputRoot:    request.OutputPath,
		Force:         request.Force,
		SourceCleanup: request.SourceCleanup,
	})
	if err != nil {
		return DecryptProjectResult{}, err
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
		return ShareSummary{}, err
	}
	return ShareSummary{
		ShareID:           result.ShareID,
		DatabaseType:      result.DatabaseType,
		FormatVersion:     result.FormatVersion,
		SchemaVersion:     result.SchemaVersion,
		TopLevelItems:     result.TopLevelItems,
		Files:             result.Files,
		Folders:           result.Folders,
		Parts:             result.Parts,
		StorageObjects:    result.StorageObjects,
		PasswordProtected: result.PasswordProtected,
	}, nil
}

func (a *App) VerifyShare(request VerifyShareRequest) (VerifyProjectResult, error) {
	result, err := a.service.VerifyShare(a.ctx, app.ShareOpen{
		DatabasePath: request.DatabasePath,
		Password:     request.Password,
	}, request.EncryptedPath)
	if err != nil {
		return VerifyProjectResult{}, err
	}
	return VerifyProjectResult{
		ProjectID:       result.ProjectID,
		CheckedObjects:  result.CheckedObjects,
		MissingObjects:  result.MissingObjects,
		TamperedObjects: result.TamperedObjects,
		ExtraObjects:    result.ExtraObjects,
		Status:          result.Status,
	}, nil
}

func (a *App) DecryptShare(request DecryptShareRequest) (DecryptShareResult, error) {
	result, err := a.service.DecryptShare(a.ctx, app.DecryptShareInput{
		DatabasePath:  request.DatabasePath,
		Password:      request.Password,
		EncryptedRoot: request.EncryptedPath,
		OutputRoot:    request.OutputPath,
		Force:         request.Force,
		SourceCleanup: request.SourceCleanup,
	})
	if err != nil {
		return DecryptShareResult{}, err
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
	result, err := a.service.ExportProject(a.ctx, app.ExportProjectInput{
		ProjectID:  request.ProjectID,
		Password:   request.Password,
		OutputPath: request.OutputPath,
		Force:      request.Force,
	})
	if err != nil {
		return ExportProjectResult{}, err
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
		return DeleteProjectResult{}, err
	}
	return DeleteProjectResult{
		ProjectID: result.ProjectID,
	}, nil
}

func (a *App) CreateProject(request CreateProjectRequest) (CreateProjectResult, error) {
	result, err := a.service.CreateProject(a.ctx, app.CreateProjectInput{
		SourcePath:     request.SourcePath,
		ContentOutput:  request.ContentOutput,
		Password:       request.Password,
		MaxPartSize:    request.MaxPartSize,
		Force:          request.Force,
		SourceCleanup:  request.SourceCleanup,
		DatabaseExport: request.DatabaseExport,
	})
	if err != nil {
		return CreateProjectResult{}, err
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
	result, err := a.service.ImportProject(a.ctx, app.ImportProjectInput{
		InputPath: request.InputPath,
		Password:  request.Password,
		Force:     request.Force,
	})
	if err != nil {
		return ImportProjectResult{}, err
	}
	return ImportProjectResult{
		ProjectID: result.ProjectID,
	}, nil
}
