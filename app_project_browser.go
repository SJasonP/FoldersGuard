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
		return ProjectBrowserState{}, frontendError(err)
	}
	return projectBrowserStateFromApp(state), nil
}

func (a *App) ApplyProjectChanges(request ApplyProjectChangesRequest) (ApplyProjectChangesResult, error) {
	renames := make([]app.ProjectRenameChange, 0, len(request.RenameChanges))
	for _, change := range request.RenameChanges {
		renames = append(renames, app.ProjectRenameChange{
			ItemPath: change.ItemPath,
			NewName:  change.NewName,
		})
	}
	moves := make([]app.ProjectMoveChange, 0, len(request.MoveChanges))
	for _, change := range request.MoveChanges {
		moves = append(moves, app.ProjectMoveChange{
			ItemPath:         change.ItemPath,
			TargetFolderPath: change.TargetFolderPath,
		})
	}
	removes := make([]app.ProjectRemoveChange, 0, len(request.RemoveChanges))
	for _, change := range request.RemoveChanges {
		removes = append(removes, app.ProjectRemoveChange{
			ItemPath: change.ItemPath,
		})
	}
	adds := make([]app.ProjectAddChange, 0, len(request.AddChanges))
	for _, change := range request.AddChanges {
		adds = append(adds, app.ProjectAddChange{
			SourcePath:       change.SourcePath,
			TargetFolderPath: change.TargetFolderPath,
			MaxPartSize:      change.MaxPartSize,
		})
	}
	createFolders := make([]app.ProjectCreateFolderChange, 0, len(request.CreateFolderChanges))
	for _, change := range request.CreateFolderChanges {
		createFolders = append(createFolders, app.ProjectCreateFolderChange{
			TargetFolderPath: change.TargetFolderPath,
			Name:             change.Name,
		})
	}
	ctx, finish := a.beginOperation("apply")
	result, err := a.service.ApplyProjectChanges(ctx, app.ApplyProjectChangesInput{
		ProjectID:           request.ProjectID,
		Password:            request.Password,
		EncryptedRoot:       request.EncryptedPath,
		RenameChanges:       renames,
		MoveChanges:         moves,
		RemoveChanges:       removes,
		AddChanges:          adds,
		CreateFolderChanges: createFolders,
	})
	finish(err)
	if err != nil {
		return ApplyProjectChangesResult{}, frontendError(err)
	}
	return ApplyProjectChangesResult{
		ProjectID:              result.ProjectID,
		AppliedRenames:         result.AppliedRenames,
		AppliedMoves:           result.AppliedMoves,
		AppliedRemoves:         result.AppliedRemoves,
		AppliedAdds:            result.AppliedAdds,
		AppliedCreatedFolders:  result.AppliedCreatedFolders,
		ManualContentGuide:     result.ManualContentGuide,
		StagedContentPath:      result.StagedContentPath,
		StagedContentName:      result.StagedContentName,
		StagedContentOnDesktop: result.StagedContentOnDesktop,
		ContentOperations:      projectContentOperationsFromApp(result.ContentOperations),
		AppliedContentChanges:  projectContentOperationsFromApp(result.AppliedContentChanges),
		BrowserState:           projectBrowserStateFromApp(result.BrowserState),
	}, nil
}

func projectContentOperationsFromApp(operations []app.ProjectContentOperation) []ProjectContentOperation {
	converted := make([]ProjectContentOperation, 0, len(operations))
	for _, operation := range operations {
		converted = append(converted, ProjectContentOperation{
			Type:       operation.Type,
			SourcePath: operation.SourcePath,
			TargetPath: operation.TargetPath,
		})
	}
	return converted
}

func projectBrowserStateFromApp(state app.ProjectBrowserState) ProjectBrowserState {
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
	}
}
