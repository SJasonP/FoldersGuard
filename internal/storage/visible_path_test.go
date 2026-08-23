package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

func TestValidateStorageVisiblePath(t *testing.T) {
	valid := uuid.New().String() + "/" + uuid.New().String()
	if err := validateStorageVisiblePath(valid); err != nil {
		t.Fatalf("valid path rejected: %v", err)
	}

	for _, value := range []string{
		"",
		"../" + uuid.New().String(),
		"/" + uuid.New().String(),
		uuid.New().String() + "//" + uuid.New().String(),
		uuid.New().String() + "\\" + uuid.New().String(),
		"ordinary-filename",
		strings.ToUpper(uuid.New().String()),
	} {
		if err := validateStorageVisiblePath(value); err == nil {
			t.Errorf("unsafe visible path %q was accepted", value)
		}
	}
}

func TestReadStorageObjectsRejectsUnsafeVisiblePath(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ApplySchema(ctx); err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO storage_objects (object_id, item_id, object_type, visible_path)
VALUES (?, ?, 'file', ?)`, uuid.New().String(), uuid.New().String(), "../escape")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.readStorageObjects(ctx); err == nil {
		t.Fatal("expected imported unsafe visible path to be rejected")
	}
}

func TestWriteStorageObjectsRejectsUnsafeVisiblePath(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ApplySchema(ctx); err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	err = writeStorageObjects(ctx, tx, []model.StorageObject{{
		ID:          uuid.New(),
		ItemID:      uuid.New(),
		Type:        model.StorageObjectTypeFile,
		VisiblePath: "../escape",
	}})
	if err == nil {
		t.Fatal("expected unsafe visible path to be rejected before storage")
	}
}
