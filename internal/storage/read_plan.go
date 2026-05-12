package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/format"
	"foldersguard/internal/fsmeta"
	"foldersguard/internal/model"
)

func (s *Store) ReadPlannedProject(ctx context.Context) (model.PlannedProject, error) {
	meta, err := s.Meta(ctx)
	if err != nil {
		return model.PlannedProject{}, err
	}
	if err := validateMeta(meta); err != nil {
		return model.PlannedProject{}, err
	}
	projectID, err := parseUUIDMeta(meta, "project_id")
	if err != nil {
		return model.PlannedProject{}, err
	}
	rootFolderID, err := parseUUIDMeta(meta, "root_folder_id")
	if err != nil {
		return model.PlannedProject{}, err
	}
	createdAt, err := parseTimeMeta(meta, "created_at")
	if err != nil {
		return model.PlannedProject{}, err
	}
	updatedAt, err := parseTimeMeta(meta, "updated_at")
	if err != nil {
		return model.PlannedProject{}, err
	}

	items, rootItem, err := s.readItems(ctx, rootFolderID)
	if err != nil {
		return model.PlannedProject{}, err
	}
	folders, rootFolder, err := s.readFolders(ctx, rootFolderID, rootItem.Type)
	if err != nil {
		return model.PlannedProject{}, err
	}
	files, err := s.readFiles(ctx)
	if err != nil {
		return model.PlannedProject{}, err
	}
	parts, err := s.readParts(ctx)
	if err != nil {
		return model.PlannedProject{}, err
	}
	objects, err := s.readStorageObjects(ctx)
	if err != nil {
		return model.PlannedProject{}, err
	}

	return model.PlannedProject{
		Project: model.Project{
			ID:           projectID,
			RootFolderID: rootFolderID,
			DatabaseType: meta["database_type"],
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		},
		RootItem:       rootItem,
		RootFolder:     rootFolder,
		Items:          items,
		Folders:        folders,
		Files:          files,
		Parts:          parts,
		StorageObjects: objects,
	}, nil
}

func validateMeta(meta map[string]string) error {
	required := map[string]string{
		"app_id":                format.AppID,
		"format_version":        format.FormatVersion,
		"crypto_suite":          format.CryptoSuite,
		"content_crypto_suite":  format.ContentAlgorithm,
		"database_crypto_suite": format.DatabaseAlgorithm,
	}
	for key, want := range required {
		if got := meta[key]; got != want {
			return fmt.Errorf("meta %s = %q, want %q", key, got, want)
		}
	}
	if meta["database_type"] == "" {
		return fmt.Errorf("meta database_type is required")
	}
	return nil
}

func (s *Store) readItems(ctx context.Context, rootFolderID uuid.UUID) ([]model.Item, model.Item, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT item_id, parent_id, item_type, visible_name, real_name,
	original_mode, original_mod_time, original_access_time, original_birth_time, windows_attributes, metadata_capabilities,
	created_at, updated_at, deleted_at
FROM items
ORDER BY parent_id IS NOT NULL, sort_name, item_id`)
	if err != nil {
		return nil, model.Item{}, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []model.Item
	var rootItem model.Item
	foundRoot := false
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, model.Item{}, err
		}
		if item.ID == rootFolderID {
			rootItem = item
			foundRoot = true
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, model.Item{}, fmt.Errorf("iterate items: %w", err)
	}
	if !foundRoot {
		return nil, model.Item{}, fmt.Errorf("root item %s not found", rootFolderID)
	}
	if rootItem.Type != model.ItemTypeFolder {
		return nil, model.Item{}, fmt.Errorf("root item %s type = %s, want folder", rootFolderID, rootItem.Type)
	}
	if rootItem.ParentID != nil {
		return nil, model.Item{}, fmt.Errorf("root item %s must not have a parent", rootFolderID)
	}
	return items, rootItem, nil
}

func scanItem(scanner interface {
	Scan(dest ...any) error
}) (model.Item, error) {
	var idText, itemType, visibleNameText, realName, originalModTimeText, metadataCapsText, createdAtText, updatedAtText string
	var originalModeValue int64
	var parentIDText, accessTimeText, birthTimeText, deletedAtText sql.NullString
	var windowsAttributes sql.NullInt64
	if err := scanner.Scan(
		&idText,
		&parentIDText,
		&itemType,
		&visibleNameText,
		&realName,
		&originalModeValue,
		&originalModTimeText,
		&accessTimeText,
		&birthTimeText,
		&windowsAttributes,
		&metadataCapsText,
		&createdAtText,
		&updatedAtText,
		&deletedAtText,
	); err != nil {
		return model.Item{}, fmt.Errorf("scan item: %w", err)
	}

	id, err := uuid.Parse(idText)
	if err != nil {
		return model.Item{}, fmt.Errorf("parse item id %q: %w", idText, err)
	}
	visibleName, err := uuid.Parse(visibleNameText)
	if err != nil {
		return model.Item{}, fmt.Errorf("parse item visible name %q: %w", visibleNameText, err)
	}
	createdAt, err := parseTime(createdAtText)
	if err != nil {
		return model.Item{}, fmt.Errorf("parse item created_at for %s: %w", id, err)
	}
	updatedAt, err := parseTime(updatedAtText)
	if err != nil {
		return model.Item{}, fmt.Errorf("parse item updated_at for %s: %w", id, err)
	}
	originalModTime, err := parseTime(originalModTimeText)
	if err != nil {
		return model.Item{}, fmt.Errorf("parse item original_mod_time for %s: %w", id, err)
	}
	if originalModeValue < 0 || originalModeValue > int64(^uint32(0)) {
		return model.Item{}, fmt.Errorf("item %s original mode out of range", id)
	}
	originalMode := uint32(originalModeValue)

	var parentID *uuid.UUID
	if parentIDText.Valid {
		parsed, err := uuid.Parse(parentIDText.String)
		if err != nil {
			return model.Item{}, fmt.Errorf("parse item parent id %q: %w", parentIDText.String, err)
		}
		parentID = &parsed
	}
	var deletedAt *time.Time
	if deletedAtText.Valid {
		parsed, err := parseTime(deletedAtText.String)
		if err != nil {
			return model.Item{}, fmt.Errorf("parse item deleted_at for %s: %w", id, err)
		}
		deletedAt = &parsed
	}
	var originalAccessTime *time.Time
	if accessTimeText.Valid {
		parsed, err := parseTime(accessTimeText.String)
		if err != nil {
			return model.Item{}, fmt.Errorf("parse item original_access_time for %s: %w", id, err)
		}
		originalAccessTime = &parsed
	}
	var originalBirthTime *time.Time
	if birthTimeText.Valid {
		parsed, err := parseTime(birthTimeText.String)
		if err != nil {
			return model.Item{}, fmt.Errorf("parse item original_birth_time for %s: %w", id, err)
		}
		originalBirthTime = &parsed
	}
	var windowsAttrs *uint32
	if windowsAttributes.Valid {
		if windowsAttributes.Int64 < 0 || windowsAttributes.Int64 > int64(^uint32(0)) {
			return model.Item{}, fmt.Errorf("item %s windows attributes out of range", id)
		}
		attrs := uint32(windowsAttributes.Int64)
		windowsAttrs = &attrs
	}

	typedItem, err := parseItemType(itemType)
	if err != nil {
		return model.Item{}, err
	}
	return model.Item{
		ID:                 id,
		ParentID:           parentID,
		Type:               typedItem,
		VisibleName:        visibleName,
		RealName:           realName,
		OriginalMode:       originalMode,
		OriginalModTime:    originalModTime,
		OriginalAccessTime: originalAccessTime,
		OriginalBirthTime:  originalBirthTime,
		WindowsAttributes:  windowsAttrs,
		MetadataCaps:       fsmeta.ParseCapabilities(metadataCapsText),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		DeletedAt:          deletedAt,
	}, nil
}

func (s *Store) readFolders(ctx context.Context, rootFolderID uuid.UUID, rootItemType model.ItemType) ([]model.Folder, model.Folder, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT folder_id, folder_key, original_size FROM folders ORDER BY folder_id`)
	if err != nil {
		return nil, model.Folder{}, fmt.Errorf("query folders: %w", err)
	}
	defer rows.Close()

	var folders []model.Folder
	var rootFolder model.Folder
	foundRoot := false
	for rows.Next() {
		var idText string
		var key []byte
		var originalSize int64
		if err := rows.Scan(&idText, &key, &originalSize); err != nil {
			return nil, model.Folder{}, fmt.Errorf("scan folder: %w", err)
		}
		id, err := uuid.Parse(idText)
		if err != nil {
			return nil, model.Folder{}, fmt.Errorf("parse folder id %q: %w", idText, err)
		}
		if len(key) != 32 {
			return nil, model.Folder{}, fmt.Errorf("folder %s key length = %d, want 32", id, len(key))
		}
		if originalSize < 0 {
			return nil, model.Folder{}, fmt.Errorf("folder %s original size is negative", id)
		}
		folder := model.Folder{ID: id, Key: key, OriginalSize: originalSize}
		if id == rootFolderID {
			rootFolder = folder
			foundRoot = true
			continue
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, model.Folder{}, fmt.Errorf("iterate folders: %w", err)
	}
	if rootItemType != model.ItemTypeFolder {
		return nil, model.Folder{}, fmt.Errorf("root item %s has unsupported type %s", rootFolderID, rootItemType)
	}
	if !foundRoot {
		return nil, model.Folder{}, fmt.Errorf("root folder %s not found", rootFolderID)
	}
	return folders, rootFolder, nil
}

func (s *Store) readFiles(ctx context.Context) ([]model.File, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT file_id, file_key, original_size, content_algorithm, storage_kind
FROM files
ORDER BY file_id`)
	if err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}
	defer rows.Close()

	var files []model.File
	for rows.Next() {
		var idText, contentAlgorithm, storageKindText string
		var key []byte
		var originalSize int64
		if err := rows.Scan(&idText, &key, &originalSize, &contentAlgorithm, &storageKindText); err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		id, err := uuid.Parse(idText)
		if err != nil {
			return nil, fmt.Errorf("parse file id %q: %w", idText, err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("file %s key length = %d, want 32", id, len(key))
		}
		if originalSize < 0 {
			return nil, fmt.Errorf("file %s original size is negative", id)
		}
		if contentAlgorithm != format.ContentAlgorithm {
			return nil, fmt.Errorf("file %s content algorithm = %q, want %q", id, contentAlgorithm, format.ContentAlgorithm)
		}
		storageKind, err := parseStorageKind(storageKindText)
		if err != nil {
			return nil, err
		}
		files = append(files, model.File{
			ID:               id,
			Key:              key,
			OriginalSize:     originalSize,
			ContentAlgorithm: contentAlgorithm,
			StorageKind:      storageKind,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate files: %w", err)
	}
	return files, nil
}

func (s *Store) readParts(ctx context.Context) ([]model.Part, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT part_id, file_id, part_index, visible_name, offset, size, integrity
FROM parts
ORDER BY file_id, part_index`)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	defer rows.Close()

	var parts []model.Part
	for rows.Next() {
		var partIDText, fileIDText, visibleNameText string
		var index int
		var offset, size int64
		var integrity []byte
		if err := rows.Scan(&partIDText, &fileIDText, &index, &visibleNameText, &offset, &size, &integrity); err != nil {
			return nil, fmt.Errorf("scan part: %w", err)
		}
		if index < 0 {
			return nil, fmt.Errorf("part %s index is negative", partIDText)
		}
		if offset < 0 {
			return nil, fmt.Errorf("part %s offset is negative", partIDText)
		}
		if size < 0 {
			return nil, fmt.Errorf("part %s size is negative", partIDText)
		}
		partID, err := uuid.Parse(partIDText)
		if err != nil {
			return nil, fmt.Errorf("parse part id %q: %w", partIDText, err)
		}
		fileID, err := uuid.Parse(fileIDText)
		if err != nil {
			return nil, fmt.Errorf("parse part file id %q: %w", fileIDText, err)
		}
		visibleName, err := uuid.Parse(visibleNameText)
		if err != nil {
			return nil, fmt.Errorf("parse part visible name %q: %w", visibleNameText, err)
		}
		parts = append(parts, model.Part{
			ID:          partID,
			FileID:      fileID,
			Index:       index,
			VisibleName: visibleName,
			Offset:      offset,
			Size:        size,
			Integrity:   integrity,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parts: %w", err)
	}
	return parts, nil
}

func (s *Store) readStorageObjects(ctx context.Context) ([]model.StorageObject, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT object_id, item_id, object_type, visible_path, size, integrity
FROM storage_objects
ORDER BY visible_path`)
	if err != nil {
		return nil, fmt.Errorf("query storage objects: %w", err)
	}
	defer rows.Close()

	var objects []model.StorageObject
	for rows.Next() {
		var objectIDText, itemIDText, objectTypeText, visiblePath string
		var size sql.NullInt64
		var integrity []byte
		if err := rows.Scan(&objectIDText, &itemIDText, &objectTypeText, &visiblePath, &size, &integrity); err != nil {
			return nil, fmt.Errorf("scan storage object: %w", err)
		}
		if visiblePath == "" {
			return nil, fmt.Errorf("storage object %s visible path is required", objectIDText)
		}
		objectID, err := uuid.Parse(objectIDText)
		if err != nil {
			return nil, fmt.Errorf("parse storage object id %q: %w", objectIDText, err)
		}
		itemID, err := uuid.Parse(itemIDText)
		if err != nil {
			return nil, fmt.Errorf("parse storage object item id %q: %w", itemIDText, err)
		}
		objectType, err := parseStorageObjectType(objectTypeText)
		if err != nil {
			return nil, err
		}
		var objectSize *int64
		if size.Valid {
			if size.Int64 < 0 {
				return nil, fmt.Errorf("storage object %s size is negative", objectID)
			}
			objectSize = &size.Int64
		}
		objects = append(objects, model.StorageObject{
			ID:          objectID,
			ItemID:      itemID,
			Type:        objectType,
			VisiblePath: visiblePath,
			Size:        objectSize,
			Integrity:   integrity,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate storage objects: %w", err)
	}
	return objects, nil
}

func parseUUIDMeta(meta map[string]string, key string) (uuid.UUID, error) {
	value := meta[key]
	if value == "" {
		return uuid.Nil, fmt.Errorf("meta %s is required", key)
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse meta %s: %w", key, err)
	}
	return parsed, nil
}

func parseTimeMeta(meta map[string]string, key string) (time.Time, error) {
	value := meta[key]
	if value == "" {
		return time.Time{}, fmt.Errorf("meta %s is required", key)
	}
	return parseTime(value)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func parseItemType(value string) (model.ItemType, error) {
	switch model.ItemType(value) {
	case model.ItemTypeFile:
		return model.ItemTypeFile, nil
	case model.ItemTypeFolder:
		return model.ItemTypeFolder, nil
	default:
		return "", fmt.Errorf("unsupported item type %q", value)
	}
}

func parseStorageKind(value string) (model.StorageKind, error) {
	switch model.StorageKind(value) {
	case model.StorageKindSingle:
		return model.StorageKindSingle, nil
	case model.StorageKindSplit:
		return model.StorageKindSplit, nil
	default:
		return "", fmt.Errorf("unsupported storage kind %q", value)
	}
}

func parseStorageObjectType(value string) (model.StorageObjectType, error) {
	switch model.StorageObjectType(value) {
	case model.StorageObjectTypeFile:
		return model.StorageObjectTypeFile, nil
	case model.StorageObjectTypeFolder:
		return model.StorageObjectTypeFolder, nil
	case model.StorageObjectTypePart:
		return model.StorageObjectTypePart, nil
	default:
		return "", fmt.Errorf("unsupported storage object type %q", value)
	}
}
