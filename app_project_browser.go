package main

import (
	"time"

	"foldersguard/internal/app"
)

func (a *App) OpenProjectBrowser(request OpenProjectBrowserRequest) (ProjectBrowserState, error) {
	state, err := a.service.OpenProjectBrowser(a.ctx, app.OpenProjectBrowserInput{
		ProjectID:     request.ProjectID,
		Password:      request.Password,
		EncryptedRoot: request.EncryptedPath,
	})
	if err != nil {
		return ProjectBrowserState{}, err
	}

	items := make([]ProjectBrowserItem, 0, len(state.Items))
	for _, item := range state.Items {
		modifiedAt := ""
		if !item.ModifiedAt.IsZero() {
			modifiedAt = item.ModifiedAt.Format(time.RFC3339)
		}
		items = append(items, ProjectBrowserItem{
			ID:               item.ID,
			ParentID:         item.ParentID,
			Path:             item.Path,
			ParentPath:       item.ParentPath,
			Name:             item.Name,
			Type:             item.Type,
			Size:             item.Size,
			ChildCount:       item.ChildCount,
			ModifiedAt:       modifiedAt,
			MetadataCaptured: item.MetadataCaptured,
			ContentAvailable: item.ContentAvailable,
		})
	}

	createdAt := ""
	if !state.CreatedAt.IsZero() {
		createdAt = state.CreatedAt.Format(time.RFC3339)
	}
	updatedAt := ""
	if !state.UpdatedAt.IsZero() {
		updatedAt = state.UpdatedAt.Format(time.RFC3339)
	}
	return ProjectBrowserState{
		ProjectID:        state.ProjectID,
		ProjectName:      state.ProjectName,
		RootFolderID:     state.RootFolderID,
		RootFolderName:   state.RootFolderName,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		Files:            state.Files,
		Folders:          state.Folders,
		Parts:            state.Parts,
		ContentConnected: state.ContentConnected,
		EncryptedPath:    state.EncryptedRoot,
		Items:            items,
	}, nil
}
