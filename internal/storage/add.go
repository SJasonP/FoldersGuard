package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

type AddResult struct {
	ProjectID       uuid.UUID
	OperationPlanID uuid.UUID
	Operations      []ContentOperation
}

func (s *Store) PrepareAdd(ctx context.Context, targetFolderPath string, addition model.PlannedProject) (model.PlannedProject, []ContentOperation, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	targetFolder, err := itemByRealPath(plan, targetFolderPath)
	if err != nil {
		return model.PlannedProject{}, nil, fmt.Errorf("target folder: %w", err)
	}
	if targetFolder.Type != model.ItemTypeFolder {
		return model.PlannedProject{}, nil, fmt.Errorf("target path is not a folder")
	}
	if siblingNameExists(plan, targetFolder.ID, addition.RootItem.ID, addition.RootItem.RealName) {
		return model.PlannedProject{}, nil, fmt.Errorf("sibling name %q already exists", addition.RootItem.RealName)
	}
	targetVisiblePath, err := visiblePathForItem(plan, targetFolder.ID)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	rootVisiblePath, err := visiblePathForItem(addition, addition.RootItem.ID)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}

	targetParentID := targetFolder.ID
	addition.RootItem.ParentID = &targetParentID
	finalRootPath := targetVisiblePath + "/" + addition.RootItem.VisibleName.String()
	for i, object := range addition.StorageObjects {
		newPath, ok := replaceVisiblePathPrefix(object.VisiblePath, rootVisiblePath, finalRootPath)
		if !ok {
			return model.PlannedProject{}, nil, fmt.Errorf("storage object %s is outside added root", object.ID)
		}
		addition.StorageObjects[i].VisiblePath = newPath
	}
	return addition, []ContentOperation{{
		Type:       "upload",
		SourcePath: finalRootPath,
		TargetPath: finalRootPath,
	}}, nil
}

func (s *Store) CommitAdd(ctx context.Context, targetFolderPath string, addition model.PlannedProject, operations []ContentOperation, now time.Time) (AddResult, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return AddResult{}, err
	}
	targetFolder, err := itemByRealPath(plan, targetFolderPath)
	if err != nil {
		return AddResult{}, fmt.Errorf("target folder: %w", err)
	}
	if targetFolder.Type != model.ItemTypeFolder {
		return AddResult{}, fmt.Errorf("target path is not a folder")
	}
	if addition.RootItem.ParentID == nil || *addition.RootItem.ParentID != targetFolder.ID {
		return AddResult{}, fmt.Errorf("added root parent does not match target folder")
	}
	if siblingNameExists(plan, targetFolder.ID, addition.RootItem.ID, addition.RootItem.RealName) {
		return AddResult{}, fmt.Errorf("sibling name %q already exists", addition.RootItem.RealName)
	}
	if len(operations) == 0 {
		return AddResult{}, fmt.Errorf("add operations are required")
	}

	updatedAt := formatTime(now.UTC())
	operationPlanID := uuid.New()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AddResult{}, fmt.Errorf("begin add item: %w", err)
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
		return AddResult{}, fmt.Errorf("insert add operation plan: %w", err)
	}
	for index, operation := range operations {
		if operation.Type != "upload" {
			return AddResult{}, fmt.Errorf("unsupported add operation %q", operation.Type)
		}
		stepID := uuid.New()
		if _, err := tx.ExecContext(ctx, `
INSERT INTO operation_steps (
	step_id, plan_id, step_index, operation_type, source_visible_path, target_visible_path, expected_integrity
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			stepID.String(),
			operationPlanID.String(),
			index,
			operation.Type,
			operation.SourcePath,
			operation.TargetPath,
			nil,
		); err != nil {
			return AddResult{}, fmt.Errorf("insert add operation step: %w", err)
		}
	}

	items := append([]model.Item{addition.RootItem}, addition.Items...)
	if err := writeItems(ctx, tx, items); err != nil {
		return AddResult{}, err
	}
	folders := append([]model.Folder(nil), addition.Folders...)
	if addition.RootItem.Type == model.ItemTypeFolder {
		folders = append([]model.Folder{addition.RootFolder}, folders...)
	}
	if err := writeFolders(ctx, tx, folders); err != nil {
		return AddResult{}, err
	}
	if err := writeFiles(ctx, tx, addition.Files); err != nil {
		return AddResult{}, err
	}
	if err := writeParts(ctx, tx, addition.Parts); err != nil {
		return AddResult{}, err
	}
	if err := writeStorageObjects(ctx, tx, addition.StorageObjects); err != nil {
		return AddResult{}, err
	}
	if err := updateFolderSizeAncestors(ctx, tx, targetFolder.ID, plannedFilesSize(addition)); err != nil {
		return AddResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value = ? WHERE key = 'updated_at'`, updatedAt); err != nil {
		return AddResult{}, fmt.Errorf("update metadata timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return AddResult{}, fmt.Errorf("add conflicts with existing metadata: %w", err)
		}
		return AddResult{}, fmt.Errorf("commit add item: %w", err)
	}
	return AddResult{
		ProjectID:       plan.Project.ID,
		OperationPlanID: operationPlanID,
		Operations:      operations,
	}, nil
}
