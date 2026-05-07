package main

import (
	"time"

	"foldersguard/internal/app"
)

func (a *App) ListShareableItems(request ListShareableItemsRequest) ([]ShareableItem, error) {
	items, err := a.service.ListShareableItems(a.ctx, app.DatabaseOpen{
		ProjectRef: request.ProjectID,
		Password:   request.Password,
	})
	if err != nil {
		return nil, err
	}

	result := make([]ShareableItem, 0, len(items))
	for _, item := range items {
		modifiedAt := ""
		if !item.ModifiedAt.IsZero() {
			modifiedAt = item.ModifiedAt.Format(time.RFC3339)
		}
		result = append(result, ShareableItem{
			ID:         item.ID,
			ParentID:   item.ParentID,
			Path:       item.Path,
			ParentPath: item.ParentPath,
			Name:       item.Name,
			Type:       item.Type,
			Size:       item.Size,
			ChildCount: item.ChildCount,
			ModifiedAt: modifiedAt,
		})
	}
	return result, nil
}

func (a *App) CreateShare(request CreateShareRequest) (CreateShareResult, error) {
	result, err := a.service.CreateShare(a.ctx, app.CreateShareInput{
		ProjectID:         request.ProjectID,
		ProjectPassword:   request.ProjectPassword,
		ItemPaths:         request.ItemPaths,
		OutputPath:        request.OutputPath,
		Force:             request.Force,
		PasswordProtected: request.PasswordProtected,
		SharePassword:     request.SharePassword,
	})
	if err != nil {
		return CreateShareResult{}, err
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
