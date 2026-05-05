package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"foldersguard/internal/model"
)

func TestInitProject(t *testing.T) {
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

	meta, err := store.Meta(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if meta["project_id"] != projectID.String() {
		t.Fatalf("project_id = %q, want %q", meta["project_id"], projectID.String())
	}
	if meta["root_folder_id"] != rootID.String() {
		t.Fatalf("root_folder_id = %q, want %q", meta["root_folder_id"], rootID.String())
	}
}

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

func TestInitProjectRejectsMissingFields(t *testing.T) {
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	err = store.InitProject(context.Background(), ProjectSpec{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func openMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}
