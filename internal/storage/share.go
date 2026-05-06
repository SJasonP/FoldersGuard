package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

type ShareSelection struct {
	SourceProjectID   uuid.UUID
	ShareID           uuid.UUID
	Plan              model.PlannedProject
	ContentOperations []ContentOperation
}

func (s *Store) SelectShare(ctx context.Context, itemPath string, now time.Time) (ShareSelection, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return ShareSelection{}, err
	}
	item, err := itemByRealPath(plan, itemPath)
	if err != nil {
		return ShareSelection{}, err
	}

	selectedIDs := subtreeItemIDs(plan, item.ID)
	selected := make(map[uuid.UUID]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = struct{}{}
	}

	now = now.UTC()
	shareID := uuid.New()
	rootID := uuid.New()
	rootVisible := uuid.New()
	virtualRoot := model.Item{
		ID:          rootID,
		Type:        model.ItemTypeFolder,
		VisibleName: rootVisible,
		RealName:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	rootFolder := model.Folder{ID: rootID, Key: make([]byte, 32)}
	sharedRoot := item
	sharedRoot.ParentID = &rootID

	sharePlan := model.PlannedProject{
		Project: model.Project{
			ID:           shareID,
			RootFolderID: rootID,
			DatabaseType: "share",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		RootItem:   virtualRoot,
		RootFolder: rootFolder,
		Items:      []model.Item{sharedRoot},
	}

	for _, candidate := range plan.Items {
		if candidate.ID == item.ID {
			continue
		}
		if _, ok := selected[candidate.ID]; ok {
			sharePlan.Items = append(sharePlan.Items, candidate)
		}
	}
	for _, folder := range plan.Folders {
		if _, ok := selected[folder.ID]; ok {
			sharePlan.Folders = append(sharePlan.Folders, folder)
		}
	}
	if _, ok := selected[plan.RootFolder.ID]; ok {
		sharePlan.Folders = append(sharePlan.Folders, plan.RootFolder)
	}
	for _, file := range plan.Files {
		if _, ok := selected[file.ID]; ok {
			sharePlan.Files = append(sharePlan.Files, file)
		}
	}
	selectedFiles := make(map[uuid.UUID]struct{})
	for _, file := range sharePlan.Files {
		selectedFiles[file.ID] = struct{}{}
	}
	for _, part := range plan.Parts {
		if _, ok := selectedFiles[part.FileID]; ok {
			sharePlan.Parts = append(sharePlan.Parts, part)
		}
	}

	rootVisiblePath, err := visiblePathForItem(plan, item.ID)
	if err != nil {
		return ShareSelection{}, err
	}
	shareVisiblePath := item.VisibleName.String()
	for _, object := range plan.StorageObjects {
		if _, ok := selected[object.ItemID]; ok {
			visiblePath, ok := replaceVisiblePathPrefix(object.VisiblePath, rootVisiblePath, shareVisiblePath)
			if !ok {
				return ShareSelection{}, fmt.Errorf("storage object %s is outside shared root", object.ID)
			}
			object.VisiblePath = visiblePath
			sharePlan.StorageObjects = append(sharePlan.StorageObjects, object)
		}
	}
	return ShareSelection{
		SourceProjectID: plan.Project.ID,
		ShareID:         shareID,
		Plan:            sharePlan,
		ContentOperations: []ContentOperation{{
			Type:       "copy",
			SourcePath: rootVisiblePath,
			TargetPath: shareVisiblePath,
		}},
	}, nil
}
