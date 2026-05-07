package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/crypto"
	"foldersguard/internal/fsmeta"
	"foldersguard/internal/model"
)

type CreateFolderResult struct {
	ProjectID       uuid.UUID
	OperationPlanID uuid.UUID
	FolderID        uuid.UUID
	Name            string
	Operations      []ContentOperation
}

type PreparedCreateFolder struct {
	ProjectID        uuid.UUID
	TargetFolderPath string
	TargetFolderID   uuid.UUID
	Item             model.Item
	Folder           model.Folder
	StorageObject    model.StorageObject
	Operation        ContentOperation
}

func (s *Store) PrepareCreateFolder(ctx context.Context, targetFolderPath, name string, now time.Time) (PreparedCreateFolder, error) {
	if err := validateRealNameSegment(name); err != nil {
		return PreparedCreateFolder{}, fmt.Errorf("invalid folder name: %w", err)
	}
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return PreparedCreateFolder{}, err
	}
	targetFolder, err := itemByRealPath(plan, targetFolderPath)
	if err != nil {
		return PreparedCreateFolder{}, fmt.Errorf("target folder: %w", err)
	}
	if targetFolder.Type != model.ItemTypeFolder {
		return PreparedCreateFolder{}, fmt.Errorf("target path is not a folder")
	}
	if siblingNameExists(plan, targetFolder.ID, uuid.Nil, name) {
		return PreparedCreateFolder{}, fmt.Errorf("sibling name %q already exists", name)
	}
	targetVisiblePath, err := visiblePathForItem(plan, targetFolder.ID)
	if err != nil {
		return PreparedCreateFolder{}, err
	}
	key, err := crypto.GenerateKey256()
	if err != nil {
		return PreparedCreateFolder{}, fmt.Errorf("generate folder key: %w", err)
	}

	createdAt := now.UTC()
	folderID := uuid.New()
	visibleName := uuid.New()
	parentID := targetFolder.ID
	item := model.Item{
		ID:              folderID,
		ParentID:        &parentID,
		Type:            model.ItemTypeFolder,
		VisibleName:     visibleName,
		RealName:        name,
		OriginalMode:    uint32(0o40755),
		OriginalModTime: createdAt,
		MetadataCaps:    []string{fsmeta.CapabilityMode, fsmeta.CapabilityModTime},
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
	folder := model.Folder{ID: folderID, Key: key}
	object := model.StorageObject{
		ID:          uuid.New(),
		ItemID:      folderID,
		Type:        model.StorageObjectTypeFolder,
		VisiblePath: targetVisiblePath + "/" + visibleName.String(),
	}
	operation := ContentOperation{
		Type:       "upload",
		SourcePath: object.VisiblePath,
		TargetPath: object.VisiblePath,
	}

	return PreparedCreateFolder{
		ProjectID:        plan.Project.ID,
		TargetFolderPath: targetFolderPath,
		TargetFolderID:   targetFolder.ID,
		Item:             item,
		Folder:           folder,
		StorageObject:    object,
		Operation:        operation,
	}, nil
}

func (s *Store) CommitCreateFolder(ctx context.Context, prepared PreparedCreateFolder, now time.Time) (CreateFolderResult, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return CreateFolderResult{}, err
	}
	targetFolder, err := itemByRealPath(plan, prepared.TargetFolderPath)
	if err != nil {
		return CreateFolderResult{}, fmt.Errorf("target folder: %w", err)
	}
	if targetFolder.Type != model.ItemTypeFolder {
		return CreateFolderResult{}, fmt.Errorf("target path is not a folder")
	}
	if targetFolder.ID != prepared.TargetFolderID {
		return CreateFolderResult{}, fmt.Errorf("target folder changed before create folder commit")
	}
	if prepared.Item.ParentID == nil || *prepared.Item.ParentID != targetFolder.ID {
		return CreateFolderResult{}, fmt.Errorf("created folder parent does not match target folder")
	}
	if siblingNameExists(plan, targetFolder.ID, prepared.Item.ID, prepared.Item.RealName) {
		return CreateFolderResult{}, fmt.Errorf("sibling name %q already exists", prepared.Item.RealName)
	}

	updatedAt := formatTime(now.UTC())
	operationPlanID := uuid.New()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CreateFolderResult{}, fmt.Errorf("begin create folder: %w", err)
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `
INSERT INTO operation_plans (plan_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?)`,
		operationPlanID.String(),
		"planned",
		updatedAt,
		updatedAt,
	); err != nil {
		return CreateFolderResult{}, fmt.Errorf("insert create folder operation plan: %w", err)
	}
	stepID := uuid.New()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO operation_steps (
	step_id, plan_id, step_index, operation_type, source_visible_path, target_visible_path, expected_integrity
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		stepID.String(),
		operationPlanID.String(),
		0,
		prepared.Operation.Type,
		prepared.Operation.SourcePath,
		prepared.Operation.TargetPath,
		nil,
	); err != nil {
		return CreateFolderResult{}, fmt.Errorf("insert create folder operation step: %w", err)
	}

	if err := writeItems(ctx, tx, []model.Item{prepared.Item}); err != nil {
		return CreateFolderResult{}, err
	}
	if err := writeFolders(ctx, tx, []model.Folder{prepared.Folder}); err != nil {
		return CreateFolderResult{}, err
	}
	if err := writeStorageObjects(ctx, tx, []model.StorageObject{prepared.StorageObject}); err != nil {
		return CreateFolderResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value = ? WHERE key = 'updated_at'`, updatedAt); err != nil {
		return CreateFolderResult{}, fmt.Errorf("update metadata timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return CreateFolderResult{}, fmt.Errorf("commit create folder: %w", err)
	}
	return CreateFolderResult{
		ProjectID:       plan.Project.ID,
		OperationPlanID: operationPlanID,
		FolderID:        prepared.Item.ID,
		Name:            prepared.Item.RealName,
		Operations:      []ContentOperation{prepared.Operation},
	}, nil
}
