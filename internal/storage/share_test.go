package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/fsmeta"
	"foldersguard/internal/model"
)

func TestSelectShareSupportsMultipleRootlessItems(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	projectID := uuid.New()
	rootID := uuid.New()
	docsID := uuid.New()
	fileAID := uuid.New()
	fileBID := uuid.New()
	rootVisible := uuid.New()
	docsVisible := uuid.New()
	fileAVisible := uuid.New()
	fileBVisible := uuid.New()
	now := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	rootParentID := rootID
	docsParentID := docsID
	fileASize := int64(10)
	fileBSize := int64(20)
	plan := model.PlannedProject{
		Project: model.Project{ID: projectID, RootFolderID: rootID, CreatedAt: now, UpdatedAt: now},
		RootItem: model.Item{
			ID:              rootID,
			Type:            model.ItemTypeFolder,
			VisibleName:     rootVisible,
			RealName:        "Root",
			OriginalMode:    uint32(0o40755),
			OriginalModTime: now,
			MetadataCaps:    []string{fsmeta.CapabilityMode, fsmeta.CapabilityModTime},
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		RootFolder: model.Folder{ID: rootID, Key: make([]byte, 32)},
		Items: []model.Item{
			{
				ID:              docsID,
				ParentID:        &rootParentID,
				Type:            model.ItemTypeFolder,
				VisibleName:     docsVisible,
				RealName:        "docs",
				OriginalMode:    uint32(0o40755),
				OriginalModTime: now,
				MetadataCaps:    []string{fsmeta.CapabilityMode, fsmeta.CapabilityModTime},
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			{
				ID:              fileAID,
				ParentID:        &docsParentID,
				Type:            model.ItemTypeFile,
				VisibleName:     fileAVisible,
				RealName:        "a.txt",
				OriginalMode:    uint32(0o100600),
				OriginalModTime: now,
				MetadataCaps:    []string{fsmeta.CapabilityMode, fsmeta.CapabilityModTime},
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			{
				ID:              fileBID,
				ParentID:        &rootParentID,
				Type:            model.ItemTypeFile,
				VisibleName:     fileBVisible,
				RealName:        "b.txt",
				OriginalMode:    uint32(0o100600),
				OriginalModTime: now,
				MetadataCaps:    []string{fsmeta.CapabilityMode, fsmeta.CapabilityModTime},
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
		Folders: []model.Folder{{ID: docsID, Key: make([]byte, 32)}},
		Files: []model.File{
			{ID: fileAID, Key: make([]byte, 32), OriginalSize: fileASize, ContentAlgorithm: "AES-256-GCM", StorageKind: model.StorageKindSingle},
			{ID: fileBID, Key: make([]byte, 32), OriginalSize: fileBSize, ContentAlgorithm: "AES-256-GCM", StorageKind: model.StorageKindSingle},
		},
		StorageObjects: []model.StorageObject{
			{ID: uuid.New(), ItemID: rootID, Type: model.StorageObjectTypeFolder, VisiblePath: rootVisible.String()},
			{ID: uuid.New(), ItemID: docsID, Type: model.StorageObjectTypeFolder, VisiblePath: rootVisible.String() + "/" + docsVisible.String()},
			{ID: uuid.New(), ItemID: fileAID, Type: model.StorageObjectTypeFile, VisiblePath: rootVisible.String() + "/" + docsVisible.String() + "/" + fileAVisible.String(), Size: &fileASize},
			{ID: uuid.New(), ItemID: fileBID, Type: model.StorageObjectTypeFile, VisiblePath: rootVisible.String() + "/" + fileBVisible.String(), Size: &fileBSize},
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
	plan, err = model.PopulateFolderSizes(plan)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.WritePlannedProject(ctx, plan); err != nil {
		t.Fatal(err)
	}

	selection, err := store.SelectShare(ctx, []string{"Root/docs", "Root/docs/a.txt", "Root/b.txt"}, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(selection.Plan.Items) != 3 {
		t.Fatalf("share items = %d, want 3", len(selection.Plan.Items))
	}
	if len(selection.Plan.Files) != 2 {
		t.Fatalf("share files = %d, want 2", len(selection.Plan.Files))
	}
	if len(selection.Plan.Folders) != 1 {
		t.Fatalf("share folders = %d, want 1", len(selection.Plan.Folders))
	}
	if selection.Plan.RootFolder.OriginalSize != fileASize+fileBSize {
		t.Fatalf("share root size = %d, want %d", selection.Plan.RootFolder.OriginalSize, fileASize+fileBSize)
	}
	if selection.Plan.Folders[0].OriginalSize != fileASize {
		t.Fatalf("share folder size = %d, want %d", selection.Plan.Folders[0].OriginalSize, fileASize)
	}
	if len(selection.ContentLocations) != 2 {
		t.Fatalf("content locations = %d, want 2", len(selection.ContentLocations))
	}
	for _, item := range selection.Plan.Items {
		if item.ID == docsID || item.ID == fileBID {
			if item.ParentID == nil || *item.ParentID != selection.Plan.RootItem.ID {
				t.Fatalf("top-level item %s parent = %v, want virtual root", item.ID, item.ParentID)
			}
		}
	}
}
