package project

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/crypto"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

type AddPlanner struct {
	MaxPartSize int64
	Now         func() time.Time
	NewUUID     func() uuid.UUID
	NewKey      func() ([]byte, error)
}

func (p AddPlanner) Plan(scan fswalk.ScanResult) (model.PlannedProject, error) {
	if scan.Root.AbsolutePath == "" {
		return model.PlannedProject{}, fmt.Errorf("scan root is required")
	}
	if p.MaxPartSize <= 0 {
		return model.PlannedProject{}, fmt.Errorf("max part size must be positive")
	}

	now := p.now().UTC()
	rootID := p.newUUID()
	rootVisibleName := p.newUUID()
	rootItem := model.Item{
		ID:          rootID,
		Type:        model.ItemType(scan.Root.Type),
		VisibleName: rootVisibleName,
		RealName:    filepath.Base(scan.Root.AbsolutePath),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	plan := model.PlannedProject{
		RootItem: rootItem,
		StorageObjects: []model.StorageObject{{
			ID:          p.newUUID(),
			ItemID:      rootID,
			Type:        storageObjectTypeForEntry(scan.Root.Type),
			VisiblePath: rootVisibleName.String(),
			Size:        sizePtr(scan.Root),
		}},
	}
	pathToID := map[string]uuid.UUID{".": rootID}
	pathToVisible := map[string]string{".": rootVisibleName.String()}

	switch scan.Root.Type {
	case fswalk.EntryTypeFolder:
		key, err := p.newKey()
		if err != nil {
			return model.PlannedProject{}, fmt.Errorf("generate folder key for %s: %w", scan.Root.AbsolutePath, err)
		}
		plan.RootFolder = model.Folder{ID: rootID, Key: key}
	case fswalk.EntryTypeFile:
		file, parts, objects, err := p.planFile(scan.Root, rootID, rootVisibleName.String())
		if err != nil {
			return model.PlannedProject{}, err
		}
		plan.Files = append(plan.Files, file)
		plan.Parts = append(plan.Parts, parts...)
		if len(objects) != 0 {
			plan.StorageObjects = objects
		}
	default:
		return model.PlannedProject{}, fmt.Errorf("unsupported root entry type %q", scan.Root.Type)
	}

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
			item.Type = model.ItemTypeFile
			file, parts, objects, err := p.planFile(entry, itemID, visiblePath)
			if err != nil {
				return model.PlannedProject{}, err
			}
			plan.Items = append(plan.Items, item)
			plan.Files = append(plan.Files, file)
			plan.Parts = append(plan.Parts, parts...)
			plan.StorageObjects = append(plan.StorageObjects, objects...)
		default:
			return model.PlannedProject{}, fmt.Errorf("unsupported entry type %q", entry.Type)
		}
	}

	return plan, nil
}

func (p AddPlanner) planFile(entry fswalk.Entry, itemID uuid.UUID, visiblePath string) (model.File, []model.Part, []model.StorageObject, error) {
	key, err := p.newKey()
	if err != nil {
		return model.File{}, nil, nil, fmt.Errorf("generate file key for %s: %w", entry.RootRelativePath, err)
	}
	splitPlan, err := model.PlanBalancedSplit(entry.Size, p.MaxPartSize)
	if err != nil {
		return model.File{}, nil, nil, fmt.Errorf("plan split for %s: %w", entry.RootRelativePath, err)
	}
	storageKind := model.StorageKindSingle
	if entry.Size > p.MaxPartSize {
		storageKind = model.StorageKindSplit
	}
	file := model.File{
		ID:               itemID,
		Key:              key,
		SourcePath:       entry.AbsolutePath,
		OriginalSize:     entry.Size,
		ContentAlgorithm: format.ContentAlgorithm,
		StorageKind:      storageKind,
	}
	if storageKind == model.StorageKindSingle {
		size := entry.Size
		return file, nil, []model.StorageObject{{
			ID:          p.newUUID(),
			ItemID:      itemID,
			Type:        model.StorageObjectTypeFile,
			VisiblePath: visiblePath,
			Size:        &size,
		}}, nil
	}

	objects := []model.StorageObject{{
		ID:          p.newUUID(),
		ItemID:      itemID,
		Type:        model.StorageObjectTypeFolder,
		VisiblePath: visiblePath,
	}}
	var parts []model.Part
	for _, span := range splitPlan.Parts {
		partVisibleName := p.newUUID()
		partPath := visiblePath + "/" + partVisibleName.String()
		size := span.Size
		partID := p.newUUID()
		parts = append(parts, model.Part{
			ID:          partID,
			FileID:      itemID,
			Index:       span.Index,
			VisibleName: partVisibleName,
			Offset:      span.Offset,
			Size:        span.Size,
		})
		objects = append(objects, model.StorageObject{
			ID:          partID,
			ItemID:      itemID,
			Type:        model.StorageObjectTypePart,
			VisiblePath: partPath,
			Size:        &size,
		})
	}
	return file, parts, objects, nil
}

func (p AddPlanner) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

func (p AddPlanner) newUUID() uuid.UUID {
	if p.NewUUID != nil {
		return p.NewUUID()
	}
	return uuid.New()
}

func (p AddPlanner) newKey() ([]byte, error) {
	if p.NewKey != nil {
		return p.NewKey()
	}
	return crypto.GenerateKey256()
}

func storageObjectTypeForEntry(entryType fswalk.EntryType) model.StorageObjectType {
	if entryType == fswalk.EntryTypeFolder {
		return model.StorageObjectTypeFolder
	}
	return model.StorageObjectTypeFile
}

func sizePtr(entry fswalk.Entry) *int64 {
	if entry.Type != fswalk.EntryTypeFile {
		return nil
	}
	size := entry.Size
	return &size
}
