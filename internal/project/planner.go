package project

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/crypto"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

type Planner struct {
	MaxPartSize int64
	Now         func() time.Time
	NewUUID     func() uuid.UUID
	NewKey      func() ([]byte, error)
}

func (p Planner) Plan(scan fswalk.ScanResult) (model.PlannedProject, error) {
	if scan.Root.AbsolutePath == "" {
		return model.PlannedProject{}, fmt.Errorf("scan root is required")
	}
	if p.MaxPartSize <= 0 {
		return model.PlannedProject{}, fmt.Errorf("max part size must be positive")
	}

	now := p.now().UTC()
	projectID := p.newUUID()
	rootID := p.newUUID()
	rootVisibleName := p.newUUID()
	rootKey, err := p.newKey()
	if err != nil {
		return model.PlannedProject{}, fmt.Errorf("generate root folder key: %w", err)
	}

	rootName := filepath.Base(scan.Root.AbsolutePath)
	rootItem := model.Item{
		ID:          rootID,
		Type:        model.ItemTypeFolder,
		VisibleName: rootVisibleName,
		RealName:    rootName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	plan := model.PlannedProject{
		Project: model.Project{
			ID:           projectID,
			RootFolderID: rootID,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		RootItem: rootItem,
		RootFolder: model.Folder{
			ID:  rootID,
			Key: rootKey,
		},
		StorageObjects: []model.StorageObject{{
			ID:          p.newUUID(),
			ItemID:      rootID,
			Type:        model.StorageObjectTypeFolder,
			VisiblePath: rootVisibleName.String(),
		}},
	}

	pathToID := map[string]uuid.UUID{".": rootID}
	pathToVisible := map[string]string{".": rootVisibleName.String()}

	entries := append([]fswalk.Entry(nil), scan.Entries...)
	sort.Slice(entries, func(i, j int) bool {
		leftDepth := pathDepth(entries[i].RootRelativePath)
		rightDepth := pathDepth(entries[j].RootRelativePath)
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return entries[i].RootRelativePath < entries[j].RootRelativePath
	})

	for _, entry := range entries {
		parentPath := parentRel(entry.RootRelativePath)
		parentID, ok := pathToID[parentPath]
		if !ok {
			return model.PlannedProject{}, fmt.Errorf("missing parent for %s", entry.RootRelativePath)
		}
		parentVisible, ok := pathToVisible[parentPath]
		if !ok {
			return model.PlannedProject{}, fmt.Errorf("missing visible parent for %s", entry.RootRelativePath)
		}

		itemID := p.newUUID()
		visibleName := p.newUUID()
		parentIDCopy := parentID
		item := model.Item{
			ID:          itemID,
			ParentID:    &parentIDCopy,
			VisibleName: visibleName,
			RealName:    filepath.Base(entry.RootRelativePath),
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		visiblePath := parentVisible + "/" + visibleName.String()
		pathToID[entry.RootRelativePath] = itemID
		pathToVisible[entry.RootRelativePath] = visiblePath

		switch entry.Type {
		case fswalk.EntryTypeFolder:
			key, err := p.newKey()
			if err != nil {
				return model.PlannedProject{}, fmt.Errorf("generate folder key for %s: %w", entry.RootRelativePath, err)
			}
			item.Type = model.ItemTypeFolder
			plan.Items = append(plan.Items, item)
			plan.Folders = append(plan.Folders, model.Folder{ID: itemID, Key: key})
			plan.StorageObjects = append(plan.StorageObjects, model.StorageObject{
				ID:          p.newUUID(),
				ItemID:      itemID,
				Type:        model.StorageObjectTypeFolder,
				VisiblePath: visiblePath,
			})

		case fswalk.EntryTypeFile:
			key, err := p.newKey()
			if err != nil {
				return model.PlannedProject{}, fmt.Errorf("generate file key for %s: %w", entry.RootRelativePath, err)
			}
			splitPlan, err := model.PlanBalancedSplit(entry.Size, p.MaxPartSize)
			if err != nil {
				return model.PlannedProject{}, fmt.Errorf("plan split for %s: %w", entry.RootRelativePath, err)
			}

			item.Type = model.ItemTypeFile
			storageKind := model.StorageKindSingle
			if entry.Size > p.MaxPartSize {
				storageKind = model.StorageKindSplit
			}
			plan.Items = append(plan.Items, item)
			plan.Files = append(plan.Files, model.File{
				ID:               itemID,
				Key:              key,
				SourcePath:       entry.AbsolutePath,
				OriginalSize:     entry.Size,
				ContentAlgorithm: format.ContentAlgorithm,
				StorageKind:      storageKind,
			})

			if storageKind == model.StorageKindSingle {
				size := entry.Size
				plan.StorageObjects = append(plan.StorageObjects, model.StorageObject{
					ID:          p.newUUID(),
					ItemID:      itemID,
					Type:        model.StorageObjectTypeFile,
					VisiblePath: visiblePath,
					Size:        &size,
				})
				break
			}

			plan.StorageObjects = append(plan.StorageObjects, model.StorageObject{
				ID:          p.newUUID(),
				ItemID:      itemID,
				Type:        model.StorageObjectTypeFolder,
				VisiblePath: visiblePath,
			})

			for _, span := range splitPlan.Parts {
				partVisibleName := p.newUUID()
				partPath := visiblePath + "/" + partVisibleName.String()
				size := span.Size
				partID := p.newUUID()
				plan.Parts = append(plan.Parts, model.Part{
					ID:          partID,
					FileID:      itemID,
					Index:       span.Index,
					VisibleName: partVisibleName,
					Offset:      span.Offset,
					Size:        span.Size,
				})
				plan.StorageObjects = append(plan.StorageObjects, model.StorageObject{
					ID:          partID,
					ItemID:      itemID,
					Type:        model.StorageObjectTypePart,
					VisiblePath: partPath,
					Size:        &size,
				})
			}

		default:
			return model.PlannedProject{}, fmt.Errorf("unsupported entry type %q", entry.Type)
		}
	}

	return plan, nil
}

func (p Planner) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

func (p Planner) newUUID() uuid.UUID {
	if p.NewUUID != nil {
		return p.NewUUID()
	}
	return uuid.New()
}

func (p Planner) newKey() ([]byte, error) {
	if p.NewKey != nil {
		return p.NewKey()
	}
	return crypto.GenerateKey256()
}

func parentRel(path string) string {
	dir := filepath.ToSlash(filepath.Dir(path))
	if dir == "." || dir == "" {
		return "."
	}
	return dir
}

func pathDepth(path string) int {
	if path == "." || path == "" {
		return 0
	}
	return strings.Count(filepath.ToSlash(path), "/") + 1
}
