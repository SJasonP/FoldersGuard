package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"foldersguard/internal/model"
)

func (s *Store) WritePlannedProject(ctx context.Context, plan model.PlannedProject) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin write planned project: %w", err)
	}
	defer rollback(tx)

	if err := writeItems(ctx, tx, append([]model.Item{plan.RootItem}, plan.Items...)); err != nil {
		return err
	}
	if err := writeFolders(ctx, tx, append([]model.Folder{plan.RootFolder}, plan.Folders...)); err != nil {
		return err
	}
	if err := writeFiles(ctx, tx, plan.Files); err != nil {
		return err
	}
	if err := writeParts(ctx, tx, plan.Parts); err != nil {
		return err
	}
	if err := writeStorageObjects(ctx, tx, plan.StorageObjects); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit write planned project: %w", err)
	}
	return nil
}

func writeItems(ctx context.Context, tx *sql.Tx, items []model.Item) error {
	for _, item := range items {
		var parentID any
		if item.ParentID != nil {
			parentID = item.ParentID.String()
		}
		var deletedAt any
		if item.DeletedAt != nil {
			deletedAt = formatTime(*item.DeletedAt)
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO items (
	item_id, parent_id, item_type, visible_name, real_name, sort_name, created_at, updated_at, deleted_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID.String(),
			parentID,
			string(item.Type),
			item.VisibleName.String(),
			item.RealName,
			sortName(item.RealName),
			formatTime(item.CreatedAt),
			formatTime(item.UpdatedAt),
			deletedAt,
		); err != nil {
			return fmt.Errorf("insert item %s: %w", item.ID, err)
		}
	}
	return nil
}

func writeFolders(ctx context.Context, tx *sql.Tx, folders []model.Folder) error {
	for _, folder := range folders {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO folders (folder_id, folder_key) VALUES (?, ?)`,
			folder.ID.String(),
			folder.Key,
		); err != nil {
			return fmt.Errorf("insert folder %s: %w", folder.ID, err)
		}
	}
	return nil
}

func writeFiles(ctx context.Context, tx *sql.Tx, files []model.File) error {
	for _, file := range files {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO files (
	file_id, file_key, original_size, content_algorithm, storage_kind
) VALUES (?, ?, ?, ?, ?)`,
			file.ID.String(),
			file.Key,
			file.OriginalSize,
			file.ContentAlgorithm,
			string(file.StorageKind),
		); err != nil {
			return fmt.Errorf("insert file %s: %w", file.ID, err)
		}
	}
	return nil
}

func writeParts(ctx context.Context, tx *sql.Tx, parts []model.Part) error {
	for _, part := range parts {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO parts (
	part_id, file_id, part_index, visible_name, offset, size, integrity
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			part.ID.String(),
			part.FileID.String(),
			part.Index,
			part.VisibleName.String(),
			part.Offset,
			part.Size,
			part.Integrity,
		); err != nil {
			return fmt.Errorf("insert part %s: %w", part.ID, err)
		}
	}
	return nil
}

func writeStorageObjects(ctx context.Context, tx *sql.Tx, objects []model.StorageObject) error {
	for _, object := range objects {
		var size any
		if object.Size != nil {
			size = *object.Size
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO storage_objects (
	object_id, item_id, object_type, visible_path, size, integrity
) VALUES (?, ?, ?, ?, ?, ?)`,
			object.ID.String(),
			object.ItemID.String(),
			string(object.Type),
			object.VisiblePath,
			size,
			object.Integrity,
		); err != nil {
			return fmt.Errorf("insert storage object %s: %w", object.ID, err)
		}
	}
	return nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
