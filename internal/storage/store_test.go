package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
