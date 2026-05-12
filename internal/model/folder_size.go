package model

import (
	"fmt"

	"github.com/google/uuid"
)

// PopulateFolderSizes updates every folder with the sum of file sizes below it.
func PopulateFolderSizes(plan PlannedProject) (PlannedProject, error) {
	fileSizeByID := make(map[uuid.UUID]int64, len(plan.Files))
	for _, file := range plan.Files {
		if file.OriginalSize < 0 {
			return PlannedProject{}, fmt.Errorf("file %s original size is negative", file.ID)
		}
		fileSizeByID[file.ID] = file.OriginalSize
	}

	folderIndexByID := make(map[uuid.UUID]int, len(plan.Folders))
	for i, folder := range plan.Folders {
		folderIndexByID[folder.ID] = i
	}

	itemByID := make(map[uuid.UUID]Item, len(plan.Items)+1)
	itemByID[plan.RootItem.ID] = plan.RootItem
	childrenByParent := make(map[uuid.UUID][]Item)
	for _, item := range plan.Items {
		if item.ParentID == nil {
			return PlannedProject{}, fmt.Errorf("non-root item %s has no parent", item.ID)
		}
		itemByID[item.ID] = item
		childrenByParent[*item.ParentID] = append(childrenByParent[*item.ParentID], item)
	}
	for parentID := range childrenByParent {
		if _, ok := itemByID[parentID]; !ok {
			return PlannedProject{}, fmt.Errorf("parent item %s not found", parentID)
		}
	}

	visiting := make(map[uuid.UUID]bool)
	visited := make(map[uuid.UUID]bool)
	var sizeOf func(Item) (int64, error)
	sizeOf = func(item Item) (int64, error) {
		if visiting[item.ID] {
			return 0, fmt.Errorf("item cycle detected at %s", item.ID)
		}
		if visited[item.ID] {
			if item.Type == ItemTypeFile {
				return fileSizeByID[item.ID], nil
			}
			if item.ID == plan.RootFolder.ID {
				return plan.RootFolder.OriginalSize, nil
			}
			index, ok := folderIndexByID[item.ID]
			if !ok {
				return 0, fmt.Errorf("folder %s not found", item.ID)
			}
			return plan.Folders[index].OriginalSize, nil
		}

		visiting[item.ID] = true
		defer delete(visiting, item.ID)
		switch item.Type {
		case ItemTypeFile:
			size, ok := fileSizeByID[item.ID]
			if !ok {
				return 0, fmt.Errorf("file %s not found", item.ID)
			}
			visited[item.ID] = true
			return size, nil
		case ItemTypeFolder:
			var total int64
			for _, child := range childrenByParent[item.ID] {
				size, err := sizeOf(child)
				if err != nil {
					return 0, err
				}
				total += size
			}
			if item.ID == plan.RootFolder.ID {
				plan.RootFolder.OriginalSize = total
			} else {
				index, ok := folderIndexByID[item.ID]
				if !ok {
					return 0, fmt.Errorf("folder %s not found", item.ID)
				}
				plan.Folders[index].OriginalSize = total
			}
			visited[item.ID] = true
			return total, nil
		default:
			return 0, fmt.Errorf("unsupported item type %q for %s", item.Type, item.ID)
		}
	}

	if plan.RootItem.Type == ItemTypeFolder {
		if plan.RootFolder.ID != plan.RootItem.ID {
			return PlannedProject{}, fmt.Errorf("root folder %s does not match root item %s", plan.RootFolder.ID, plan.RootItem.ID)
		}
		if _, err := sizeOf(plan.RootItem); err != nil {
			return PlannedProject{}, err
		}
	}
	return plan, nil
}
