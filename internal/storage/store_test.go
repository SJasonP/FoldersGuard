package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
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
	rootKey := make([]byte, 32)
	createdAt := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)

	if err := store.InitProject(ctx, ProjectSpec{
		ProjectID:       projectID,
		RootFolderID:    rootID,
		RootVisibleName: visibleName,
		RootRealName:    "Root",
		RootFolderKey:   rootKey,
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

	var itemType, realName string
	err = db.QueryRowContext(ctx, `
SELECT item_type, real_name FROM items WHERE item_id = ?`,
		rootID.String(),
	).Scan(&itemType, &realName)
	if err != nil {
		t.Fatal(err)
	}
	if itemType != "folder" || realName != "Root" {
		t.Fatalf("root item = (%s, %s), want (folder, Root)", itemType, realName)
	}

	var keyLen int
	err = db.QueryRowContext(ctx, `
SELECT length(folder_key) FROM folders WHERE folder_id = ?`,
		rootID.String(),
	).Scan(&keyLen)
	if err != nil {
		t.Fatal(err)
	}
	if keyLen != 32 {
		t.Fatalf("folder key length = %d, want 32", keyLen)
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
