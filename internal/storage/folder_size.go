package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

func plannedFilesSize(plan model.PlannedProject) int64 {
	var total int64
	for _, file := range plan.Files {
		total += file.OriginalSize
	}
	return total
}

func itemOriginalSize(plan model.PlannedProject, item model.Item) (int64, error) {
	switch item.Type {
	case model.ItemTypeFile:
		for _, file := range plan.Files {
			if file.ID == item.ID {
				return file.OriginalSize, nil
			}
		}
		return 0, fmt.Errorf("file %s not found", item.ID)
	case model.ItemTypeFolder:
		if item.ID == plan.RootFolder.ID {
			return plan.RootFolder.OriginalSize, nil
		}
		for _, folder := range plan.Folders {
			if folder.ID == item.ID {
				return folder.OriginalSize, nil
			}
		}
		return 0, fmt.Errorf("folder %s not found", item.ID)
	default:
		return 0, fmt.Errorf("unsupported item type %q for %s", item.Type, item.ID)
	}
}

func updateFolderSizeAncestors(ctx context.Context, tx *sql.Tx, startID uuid.UUID, delta int64) error {
	if delta == 0 {
		return nil
	}

	result, err := tx.ExecContext(ctx, `
WITH RECURSIVE ancestors(folder_id) AS (
	SELECT ?
	UNION ALL
	SELECT items.parent_id
	FROM items
	JOIN ancestors ON items.item_id = ancestors.folder_id
	WHERE items.parent_id IS NOT NULL
)
UPDATE folders
SET original_size = original_size + ?
WHERE folder_id IN (SELECT folder_id FROM ancestors)`,
		startID.String(),
		delta,
	)
	if err != nil {
		return fmt.Errorf("update ancestor folder sizes from %s: %w", startID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check ancestor folder size updates from %s: %w", startID, err)
	}
	if affected == 0 {
		return fmt.Errorf("folder %s not found while updating size", startID)
	}
	return nil
}
