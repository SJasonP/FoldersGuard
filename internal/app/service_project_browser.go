package app

import (
	"context"
	"fmt"
	"sort"

	"foldersguard/internal/project"
)

func (s Service) OpenProjectBrowser(ctx context.Context, input OpenProjectBrowserInput) (ProjectBrowserState, error) {
	plan, meta, err := s.ReadDatabase(ctx, DatabaseOpen{
		ProjectRef: input.ProjectID,
		Password:   input.Password,
	})
	if err != nil {
		return ProjectBrowserState{}, err
	}
	if meta["database_type"] != "project" {
		return ProjectBrowserState{}, fmt.Errorf("database type = %q, want project", meta["database_type"])
	}

	contentConnected := false
	connectedPath := ""
	available := map[string]bool{}
	if input.EncryptedRoot != "" {
		if err := ValidateExistingDirectory(input.EncryptedRoot, "content"); err != nil {
			return ProjectBrowserState{}, err
		}
		match, err := project.MatchAvailableContent(ctx, input.EncryptedRoot, plan)
		if err != nil {
			return ProjectBrowserState{}, err
		}
		contentConnected = true
		connectedPath = input.EncryptedRoot
		for _, file := range plan.Files {
			available[file.ID.String()] = match.HasFile(file.ID.String())
		}
		for _, folder := range plan.Folders {
			available[folder.ID.String()] = match.HasFolder(folder.ID.String())
		}
		available[plan.RootItem.ID.String()] = match.HasFolder(plan.RootItem.ID.String())
	}

	items, err := shareableItems(plan)
	if err != nil {
		return ProjectBrowserState{}, err
	}
	browserItems := make([]ProjectBrowserItem, 0, len(items))
	for _, item := range items {
		browserItems = append(browserItems, ProjectBrowserItem{
			ID:               item.ID,
			ParentID:         item.ParentID,
			Path:             item.Path,
			ParentPath:       item.ParentPath,
			Name:             item.Name,
			Type:             item.Type,
			Size:             item.Size,
			ChildCount:       item.ChildCount,
			ModifiedAt:       item.ModifiedAt,
			MetadataCaptured: true,
			ContentAvailable: !contentConnected || available[item.ID],
		})
	}
	sort.Slice(browserItems, func(i, j int) bool {
		return browserItems[i].Path < browserItems[j].Path
	})

	return ProjectBrowserState{
		ProjectID:        plan.Project.ID.String(),
		ProjectName:      plan.RootItem.RealName,
		RootFolderID:     plan.RootFolder.ID.String(),
		RootFolderName:   plan.RootItem.RealName,
		CreatedAt:        plan.Project.CreatedAt.UTC(),
		UpdatedAt:        plan.Project.UpdatedAt.UTC(),
		Files:            len(plan.Files),
		Folders:          CountFolders(plan),
		Parts:            len(plan.Parts),
		ContentConnected: contentConnected,
		EncryptedRoot:    connectedPath,
		Items:            browserItems,
	}, nil
}
