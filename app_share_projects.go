package main

import "foldersguard/internal/app"

func (a *App) CreateShare(request CreateShareRequest) (CreateShareResult, error) {
	ctx, finish := a.beginOperation("share")
	result, err := a.service.CreateShare(ctx, app.CreateShareInput{
		ProjectID:         request.ProjectID,
		ProjectPassword:   request.ProjectPassword,
		ItemPaths:         request.ItemPaths,
		OutputPath:        request.OutputPath,
		Force:             request.Force,
		PasswordProtected: request.PasswordProtected,
		SharePassword:     request.SharePassword,
	})
	finish(err)
	if err != nil {
		return CreateShareResult{}, frontendError(err)
	}

	locations := make([]ShareContentLocation, 0, len(result.ContentLocations))
	for _, location := range result.ContentLocations {
		locations = append(locations, ShareContentLocation{
			SourcePath: location.SourcePath,
			TargetPath: location.TargetPath,
		})
	}
	return CreateShareResult{
		ProjectID:         result.ProjectID,
		ShareID:           result.ShareID,
		OutputPath:        result.OutputPath,
		TopLevelItems:     result.TopLevelItems,
		Files:             result.Files,
		Folders:           result.Folders,
		Parts:             result.Parts,
		PasswordProtected: result.PasswordProtected,
		ContentLocations:  locations,
	}, nil
}
