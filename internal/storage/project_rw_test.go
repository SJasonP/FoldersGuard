package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

func TestWritePlannedProject(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ApplySchema(ctx); err != nil {
		t.Fatal(err)
	}

	projectID := uuid.New()
	rootID := uuid.New()
	fileID := uuid.New()
	partID := uuid.New()
	now := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	rootVisible := uuid.New()
	fileVisible := uuid.New()
	partVisible := uuid.New()
	parentID := rootID
	size := int64(10)

	plan := model.PlannedProject{
		Project: model.Project{ID: projectID, RootFolderID: rootID, CreatedAt: now, UpdatedAt: now},
		RootItem: model.Item{
			ID:          rootID,
			Type:        model.ItemTypeFolder,
			VisibleName: rootVisible,
			RealName:    "Root",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		RootFolder: model.Folder{ID: rootID, Key: make([]byte, 32)},
		Items: []model.Item{{
			ID:          fileID,
			ParentID:    &parentID,
			Type:        model.ItemTypeFile,
			VisibleName: fileVisible,
			RealName:    "file.txt",
			CreatedAt:   now,
			UpdatedAt:   now,
		}},
		Files: []model.File{{
			ID:               fileID,
			Key:              make([]byte, 32),
			SourcePath:       "/tmp/file.txt",
			OriginalSize:     10,
			ContentAlgorithm: "AES-256-GCM",
			StorageKind:      model.StorageKindSingle,
		}},
		Parts: []model.Part{{
			ID:          partID,
			FileID:      fileID,
			Index:       0,
			VisibleName: partVisible,
			Offset:      0,
			Size:        10,
		}},
		StorageObjects: []model.StorageObject{{
			ID:          uuid.New(),
			ItemID:      fileID,
			Type:        model.StorageObjectTypeFile,
			VisiblePath: rootVisible.String() + "/" + fileVisible.String(),
			Size:        &size,
		}},
	}

	if err := store.WritePlannedProject(ctx, plan); err != nil {
		t.Fatal(err)
	}

	var itemType, realName string
	err = db.QueryRowContext(ctx, `
SELECT item_type, real_name FROM items WHERE item_id = ?`,
		fileID.String(),
	).Scan(&itemType, &realName)
	if err != nil {
		t.Fatal(err)
	}
	if itemType != "file" || realName != "file.txt" {
		t.Fatalf("file item = (%s, %s), want (file, file.txt)", itemType, realName)
	}

	var keyLen int
	err = db.QueryRowContext(ctx, `
SELECT length(file_key) FROM files WHERE file_id = ?`,
		fileID.String(),
	).Scan(&keyLen)
	if err != nil {
		t.Fatal(err)
	}
	if keyLen != 32 {
		t.Fatalf("file key length = %d, want 32", keyLen)
	}
}

func TestReadPlannedProject(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ApplySchema(ctx); err != nil {
		t.Fatal(err)
	}

	projectID := uuid.New()
	rootID := uuid.New()
	folderID := uuid.New()
	fileID := uuid.New()
	partID := uuid.New()
	rootVisible := uuid.New()
	folderVisible := uuid.New()
	fileVisible := uuid.New()
	partVisible := uuid.New()
	now := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	rootParent := rootID
	fileSize := int64(42)
	partSize := int64(21)

	plan := model.PlannedProject{
		Project: model.Project{ID: projectID, RootFolderID: rootID, CreatedAt: now, UpdatedAt: now},
		RootItem: model.Item{
			ID:          rootID,
			Type:        model.ItemTypeFolder,
			VisibleName: rootVisible,
			RealName:    "Root",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		RootFolder: model.Folder{ID: rootID, Key: bytesOf(1, 32)},
		Items: []model.Item{
			{
				ID:          folderID,
				ParentID:    &rootParent,
				Type:        model.ItemTypeFolder,
				VisibleName: folderVisible,
				RealName:    "Folder",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          fileID,
				ParentID:    &rootParent,
				Type:        model.ItemTypeFile,
				VisibleName: fileVisible,
				RealName:    "file.txt",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Folders: []model.Folder{{ID: folderID, Key: bytesOf(2, 32)}},
		Files: []model.File{{
			ID:               fileID,
			Key:              bytesOf(3, 32),
			SourcePath:       "/tmp/file.txt",
			OriginalSize:     42,
			ContentAlgorithm: "AES-256-GCM",
			StorageKind:      model.StorageKindSplit,
		}},
		Parts: []model.Part{{
			ID:          partID,
			FileID:      fileID,
			Index:       0,
			VisibleName: partVisible,
			Offset:      0,
			Size:        21,
			Integrity:   []byte{4, 5, 6},
		}},
		StorageObjects: []model.StorageObject{
			{
				ID:          uuid.New(),
				ItemID:      rootID,
				Type:        model.StorageObjectTypeFolder,
				VisiblePath: rootVisible.String(),
			},
			{
				ID:          uuid.New(),
				ItemID:      folderID,
				Type:        model.StorageObjectTypeFolder,
				VisiblePath: rootVisible.String() + "/" + folderVisible.String(),
			},
			{
				ID:          uuid.New(),
				ItemID:      fileID,
				Type:        model.StorageObjectTypeFolder,
				VisiblePath: rootVisible.String() + "/" + fileVisible.String(),
				Size:        &fileSize,
			},
			{
				ID:          partID,
				ItemID:      fileID,
				Type:        model.StorageObjectTypePart,
				VisiblePath: rootVisible.String() + "/" + fileVisible.String() + "/" + partVisible.String(),
				Size:        &partSize,
				Integrity:   []byte{7, 8, 9},
			},
		},
	}

	if err := store.InitProject(ctx, ProjectSpec{
		ProjectID:       projectID,
		RootFolderID:    rootID,
		RootVisibleName: rootVisible,
		RootRealName:    "Root",
		RootFolderKey:   plan.RootFolder.Key,
		DatabaseType:    "project",
		CreatedAt:       now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.WritePlannedProject(ctx, plan); err != nil {
		t.Fatal(err)
	}

	read, err := store.ReadPlannedProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if read.Project.ID != projectID {
		t.Fatalf("project id = %s, want %s", read.Project.ID, projectID)
	}
	if read.RootItem.ID != rootID || read.RootItem.RealName != "Root" {
		t.Fatalf("root item = (%s, %s), want (%s, Root)", read.RootItem.ID, read.RootItem.RealName, rootID)
	}
	if len(read.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(read.Items))
	}
	if len(read.Folders) != 1 {
		t.Fatalf("folders = %d, want 1", len(read.Folders))
	}
	if len(read.Files) != 1 {
		t.Fatalf("files = %d, want 1", len(read.Files))
	}
	if string(read.Files[0].Key) != string(bytesOf(3, 32)) {
		t.Fatal("file key did not round-trip")
	}
	if read.Files[0].StorageKind != model.StorageKindSplit {
		t.Fatalf("storage kind = %s, want split", read.Files[0].StorageKind)
	}
	if len(read.Parts) != 1 || read.Parts[0].VisibleName != partVisible {
		t.Fatal("part did not round-trip")
	}
	if len(read.StorageObjects) != 4 {
		t.Fatalf("storage objects = %d, want 4", len(read.StorageObjects))
	}
}

func TestReadPlannedProjectRejectsWrongAppID(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	projectID := uuid.New()
	rootID := uuid.New()
	visibleName := uuid.New()
	createdAt := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	if err := store.InitProject(ctx, ProjectSpec{
		ProjectID:       projectID,
		RootFolderID:    rootID,
		RootVisibleName: visibleName,
		RootRealName:    "Root",
		RootFolderKey:   make([]byte, 32),
		DatabaseType:    "project",
		CreatedAt:       createdAt,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE meta SET value = ? WHERE key = ?`, "wrong.app", "app_id"); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReadPlannedProject(ctx); err == nil {
		t.Fatal("expected meta validation error")
	}
}
