package app

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/model"
	"foldersguard/internal/storage"
)

func (s Service) ListShareableItems(ctx context.Context, input DatabaseOpen) ([]ShareableItem, error) {
	plan, meta, err := s.ReadDatabase(ctx, input)
	if err != nil {
		return nil, err
	}
	if meta["database_type"] != "project" {
		return nil, fmt.Errorf("database type = %q, want project", meta["database_type"])
	}
	return shareableItems(plan)
}

func (s Service) CreateShare(ctx context.Context, input CreateShareInput) (CreateShareResult, error) {
	if len(input.ItemPaths) == 0 {
		return CreateShareResult{}, fmt.Errorf("at least one item path is required")
	}
	if !format.IsSetExtension(input.OutputPath) {
		return CreateShareResult{}, fmt.Errorf("share database output must use %s extension", format.SetExtension)
	}
	if err := PrepareFileOutput(input.OutputPath, input.Force); err != nil {
		return CreateShareResult{}, err
	}

	sharePassword := input.SharePassword
	if !input.PasswordProtected {
		sharePassword = db.UnprotectedSharePassword
	} else if sharePassword == "" {
		return CreateShareResult{}, fmt.Errorf("share password is required")
	}

	projectDatabase, err := s.ActiveProjectDatabasePath(input.ProjectID)
	if err != nil {
		return CreateShareResult{}, err
	}
	database, err := db.OpenProject(ctx, db.Config{
		Path:       projectDatabase,
		DriverName: db.SQLCipherDriver,
		Password:   input.ProjectPassword,
	})
	if err != nil {
		return CreateShareResult{}, err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return CreateShareResult{}, err
	}
	selection, err := store.SelectShare(ctx, input.ItemPaths, time.Now())
	if err != nil {
		return CreateShareResult{}, err
	}
	if err := WriteShareDatabase(ctx, db.Config{
		Path:       input.OutputPath,
		DriverName: db.SQLCipherDriver,
		Password:   sharePassword,
	}, selection.Plan); err != nil {
		return CreateShareResult{}, err
	}

	locations := make([]ShareContentLocation, 0, len(selection.ContentLocations))
	for _, location := range selection.ContentLocations {
		locations = append(locations, ShareContentLocation{
			SourcePath: location.SourcePath,
			TargetPath: location.TargetPath,
		})
	}
	return CreateShareResult{
		ProjectID:         selection.SourceProjectID.String(),
		ShareID:           selection.ShareID.String(),
		OutputPath:        input.OutputPath,
		TopLevelItems:     countTopLevelItems(selection.Plan),
		Files:             len(selection.Plan.Files),
		Folders:           CountFolders(selection.Plan),
		Parts:             len(selection.Plan.Parts),
		PasswordProtected: input.PasswordProtected,
		ContentLocations:  locations,
	}, nil
}

func shareableItems(plan model.PlannedProject) ([]ShareableItem, error) {
	paths, err := projectRealPaths(plan)
	if err != nil {
		return nil, err
	}
	sizeByID := make(map[string]int64)
	for _, file := range plan.Files {
		sizeByID[file.ID.String()] = file.OriginalSize
	}
	sizeByID[plan.RootFolder.ID.String()] = plan.RootFolder.OriginalSize
	for _, folder := range plan.Folders {
		sizeByID[folder.ID.String()] = folder.OriginalSize
	}
	childCountByID := make(map[string]int)
	for _, item := range plan.Items {
		if item.ParentID != nil {
			childCountByID[item.ParentID.String()]++
		}
	}

	items := make([]ShareableItem, 0, len(plan.Items)+1)
	items = append(items, shareableItemFromModel(plan.RootItem, "", paths[plan.RootItem.ID.String()], sizeByID, childCountByID))
	for _, item := range plan.Items {
		parentPath := ""
		if item.ParentID != nil {
			parentPath = paths[item.ParentID.String()]
		}
		items = append(items, shareableItemFromModel(item, parentPath, paths[item.ID.String()], sizeByID, childCountByID))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Path < items[j].Path
	})
	return items, nil
}

func shareableItemFromModel(item model.Item, parentPath, path string, sizeByID map[string]int64, childCountByID map[string]int) ShareableItem {
	parentID := ""
	if item.ParentID != nil {
		parentID = item.ParentID.String()
	}
	return ShareableItem{
		ID:         item.ID.String(),
		ParentID:   parentID,
		Path:       path,
		ParentPath: parentPath,
		Name:       item.RealName,
		Type:       string(item.Type),
		Size:       sizeByID[item.ID.String()],
		ChildCount: childCountByID[item.ID.String()],
		ModifiedAt: item.OriginalModTime.UTC(),
	}
}

func projectRealPaths(plan model.PlannedProject) (map[string]string, error) {
	paths := map[string]string{
		plan.RootItem.ID.String(): plan.RootItem.RealName,
	}
	itemsByParent := make(map[string][]model.Item)
	for _, item := range plan.Items {
		if item.ParentID == nil {
			return nil, fmt.Errorf("non-root item %s has no parent", item.ID)
		}
		itemsByParent[item.ParentID.String()] = append(itemsByParent[item.ParentID.String()], item)
	}

	var walk func(parentID string) error
	walk = func(parentID string) error {
		children := itemsByParent[parentID]
		sort.Slice(children, func(i, j int) bool {
			return children[i].RealName < children[j].RealName
		})
		for _, item := range children {
			parentPath := paths[parentID]
			paths[item.ID.String()] = filepath.ToSlash(filepath.Join(parentPath, item.RealName))
			if err := walk(item.ID.String()); err != nil {
				return err
			}
		}
		delete(itemsByParent, parentID)
		return nil
	}
	if err := walk(plan.RootItem.ID.String()); err != nil {
		return nil, err
	}
	if len(itemsByParent) != 0 {
		return nil, fmt.Errorf("items contain missing or cyclic parent references")
	}
	return paths, nil
}
