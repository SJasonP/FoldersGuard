package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

func TestRenameItem(t *testing.T) {
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
	rootVisible := uuid.New()
	fileVisible := uuid.New()
	now := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	parentID := rootID

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
			RealName:    "old.txt",
			CreatedAt:   now,
			UpdatedAt:   now,
		}},
		Files: []model.File{{
			ID:               fileID,
			Key:              make([]byte, 32),
			SourcePath:       "/tmp/old.txt",
			OriginalSize:     10,
			ContentAlgorithm: "AES-256-GCM",
			StorageKind:      model.StorageKindSingle,
		}},
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

	result, err := store.RenameItem(ctx, "Root/old.txt", "new.txt", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != projectID || result.ItemID != fileID || result.OldName != "old.txt" || result.NewName != "new.txt" {
		t.Fatalf("rename result = %+v", result)
	}
	read, err := store.ReadPlannedProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Items) != 1 || read.Items[0].RealName != "new.txt" {
		t.Fatalf("renamed item = %+v", read.Items)
	}
}

func TestRemoveItem(t *testing.T) {
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
	rootVisible := uuid.New()
	folderVisible := uuid.New()
	fileVisible := uuid.New()
	now := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	rootParentID := rootID
	folderParentID := folderID
	fileSize := int64(10)

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
		Items: []model.Item{
			{
				ID:          folderID,
				ParentID:    &rootParentID,
				Type:        model.ItemTypeFolder,
				VisibleName: folderVisible,
				RealName:    "docs",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          fileID,
				ParentID:    &folderParentID,
				Type:        model.ItemTypeFile,
				VisibleName: fileVisible,
				RealName:    "note.txt",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Folders: []model.Folder{{ID: folderID, Key: make([]byte, 32)}},
		Files: []model.File{{
			ID:               fileID,
			Key:              make([]byte, 32),
			SourcePath:       "/tmp/note.txt",
			OriginalSize:     fileSize,
			ContentAlgorithm: "AES-256-GCM",
			StorageKind:      model.StorageKindSingle,
		}},
		StorageObjects: []model.StorageObject{
			{ID: uuid.New(), ItemID: rootID, Type: model.StorageObjectTypeFolder, VisiblePath: rootVisible.String()},
			{ID: uuid.New(), ItemID: folderID, Type: model.StorageObjectTypeFolder, VisiblePath: rootVisible.String() + "/" + folderVisible.String()},
			{ID: uuid.New(), ItemID: fileID, Type: model.StorageObjectTypeFile, VisiblePath: rootVisible.String() + "/" + folderVisible.String() + "/" + fileVisible.String(), Size: &fileSize},
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

	result, err := store.RemoveItem(ctx, "Root/docs", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != projectID || len(result.Operations) != 1 || result.Operations[0].Type != "delete" {
		t.Fatalf("remove result = %+v", result)
	}
	if result.Operations[0].TargetPath != rootVisible.String()+"/"+folderVisible.String() {
		t.Fatalf("delete target = %q", result.Operations[0].TargetPath)
	}
	read, err := store.ReadPlannedProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Items) != 0 || len(read.Folders) != 0 || len(read.Files) != 0 || len(read.StorageObjects) != 1 {
		t.Fatalf("removed project = %+v", read)
	}
}
