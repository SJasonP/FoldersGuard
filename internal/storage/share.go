package storage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

type ShareSelection struct {
	SourceProjectID  uuid.UUID
	ShareID          uuid.UUID
	Plan             model.PlannedProject
	ContentLocations []ShareContentLocation
}

type ShareContentLocation struct {
	SourcePath string
	TargetPath string
}

func (s *Store) SelectShare(ctx context.Context, itemPaths []string, now time.Time) (ShareSelection, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return ShareSelection{}, err
	}
	if len(itemPaths) == 0 {
		return ShareSelection{}, fmt.Errorf("at least one item path is required")
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
	}

	selected, topLevelIDs, err := selectedShareItems(plan, itemPaths)
	if err != nil {
		return ShareSelection{}, err
	}
	topLevelSet := itemIDSet(topLevelIDs)

	for _, item := range topLevelIDs {
		item.ParentID = &rootID
		sharePlan.Items = append(sharePlan.Items, item)
	}
	for _, candidate := range plan.Items {
		if _, topLevel := topLevelSet[candidate.ID]; topLevel {
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

	visibleRewrites, err := shareVisiblePathRewrites(plan, topLevelIDs)
	if err != nil {
		return ShareSelection{}, err
	}
	contentLocations := make([]ShareContentLocation, 0, len(visibleRewrites))
	for _, rewrite := range visibleRewrites {
		contentLocations = append(contentLocations, ShareContentLocation{
			SourcePath: rewrite.source,
			TargetPath: rewrite.target,
		})
	}
	for _, object := range plan.StorageObjects {
		if _, ok := selected[object.ItemID]; ok {
			visiblePath, ok := rewriteShareVisiblePath(object.VisiblePath, visibleRewrites)
			if !ok {
				return ShareSelection{}, fmt.Errorf("storage object %s is outside selected roots", object.ID)
			}
			object.VisiblePath = visiblePath
			sharePlan.StorageObjects = append(sharePlan.StorageObjects, object)
		}
	}
	return ShareSelection{
		SourceProjectID:  plan.Project.ID,
		ShareID:          shareID,
		Plan:             sharePlan,
		ContentLocations: contentLocations,
	}, nil
}

func selectedShareItems(plan model.PlannedProject, itemPaths []string) (map[uuid.UUID]struct{}, []model.Item, error) {
	selected := make(map[uuid.UUID]struct{})
	topLevelByID := make(map[uuid.UUID]model.Item)
	for _, itemPath := range itemPaths {
		item, err := itemByRealPath(plan, itemPath)
		if err != nil {
			return nil, nil, err
		}
		for _, id := range subtreeItemIDs(plan, item.ID) {
			selected[id] = struct{}{}
		}
		topLevelByID[item.ID] = item
	}

	for id := range topLevelByID {
		for otherID := range topLevelByID {
			if id == otherID {
				continue
			}
			if isDescendantOf(plan, id, otherID) {
				delete(topLevelByID, id)
				break
			}
		}
	}

	topLevel := make([]model.Item, 0, len(topLevelByID))
	for _, item := range topLevelByID {
		topLevel = append(topLevel, item)
	}
	sort.Slice(topLevel, func(i, j int) bool {
		left, _ := visiblePathForItem(plan, topLevel[i].ID)
		right, _ := visiblePathForItem(plan, topLevel[j].ID)
		return left < right
	})
	return selected, topLevel, nil
}

func itemIDSet(items []model.Item) map[uuid.UUID]struct{} {
	ids := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		ids[item.ID] = struct{}{}
	}
	return ids
}

type visibleRewrite struct {
	source string
	target string
}

func shareVisiblePathRewrites(plan model.PlannedProject, topLevel []model.Item) ([]visibleRewrite, error) {
	rewrites := make([]visibleRewrite, 0, len(topLevel))
	for _, item := range topLevel {
		source, err := visiblePathForItem(plan, item.ID)
		if err != nil {
			return nil, err
		}
		rewrites = append(rewrites, visibleRewrite{
			source: source,
			target: item.VisibleName.String(),
		})
	}
	sort.Slice(rewrites, func(i, j int) bool {
		return len(rewrites[i].source) > len(rewrites[j].source)
	})
	return rewrites, nil
}

func rewriteShareVisiblePath(path string, rewrites []visibleRewrite) (string, bool) {
	for _, rewrite := range rewrites {
		if visiblePath, ok := replaceVisiblePathPrefix(path, rewrite.source, rewrite.target); ok {
			return visiblePath, true
		}
	}
	return "", false
}
