package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

type MoveResult struct {
	ProjectID       uuid.UUID
	OperationPlanID uuid.UUID
	Operations      []ContentOperation
}

func (s *Store) MoveItem(ctx context.Context, itemPath, targetFolderPath string, now time.Time) (MoveResult, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return MoveResult{}, err
	}
	item, err := itemByRealPath(plan, itemPath)
	if err != nil {
		return MoveResult{}, err
	}
	if item.ParentID == nil {
		return MoveResult{}, fmt.Errorf("root item cannot be moved")
	}
	targetFolder, err := itemByRealPath(plan, targetFolderPath)
	if err != nil {
		return MoveResult{}, fmt.Errorf("target folder: %w", err)
	}
	if targetFolder.Type != model.ItemTypeFolder {
		return MoveResult{}, fmt.Errorf("target path is not a folder")
	}
	if targetFolder.ID == item.ID || isDescendantOf(plan, targetFolder.ID, item.ID) {
		return MoveResult{}, fmt.Errorf("target folder cannot be inside moved item")
	}
	if siblingNameExists(plan, targetFolder.ID, item.ID, item.RealName) {
		return MoveResult{}, fmt.Errorf("sibling name %q already exists", item.RealName)
	}

	sourcePrefix, err := visiblePathForItem(plan, item.ID)
	if err != nil {
		return MoveResult{}, err
	}
	targetParentPath, err := visiblePathForItem(plan, targetFolder.ID)
	if err != nil {
		return MoveResult{}, err
	}
	targetPrefix := targetParentPath + "/" + item.VisibleName.String()
	operation := ContentOperation{
		Type:       "move",
		SourcePath: sourcePrefix,
		TargetPath: targetPrefix,
	}

	updatedAt := formatTime(now.UTC())
	operationPlanID := uuid.New()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MoveResult{}, fmt.Errorf("begin move item: %w", err)
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
		return MoveResult{}, fmt.Errorf("insert move operation plan: %w", err)
	}
	stepID := uuid.New()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO operation_steps (
	step_id, plan_id, step_index, operation_type, source_visible_path, target_visible_path, expected_integrity
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		stepID.String(),
		operationPlanID.String(),
		0,
		operation.Type,
		operation.SourcePath,
		operation.TargetPath,
		nil,
	); err != nil {
		return MoveResult{}, fmt.Errorf("insert move operation step: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE items
SET parent_id = ?, updated_at = ?
WHERE item_id = ?`,
		targetFolder.ID.String(),
		updatedAt,
		item.ID.String(),
	); err != nil {
		return MoveResult{}, fmt.Errorf("move item %s: %w", item.ID, err)
	}

	for _, object := range plan.StorageObjects {
		newPath, ok := replaceVisiblePathPrefix(object.VisiblePath, sourcePrefix, targetPrefix)
		if !ok {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE storage_objects
SET visible_path = ?
WHERE object_id = ?`,
			newPath,
			object.ID.String(),
		); err != nil {
			return MoveResult{}, fmt.Errorf("move storage object %s: %w", object.ID, err)
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value = ? WHERE key = 'updated_at'`, updatedAt); err != nil {
		return MoveResult{}, fmt.Errorf("update metadata timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return MoveResult{}, fmt.Errorf("commit move item: %w", err)
	}

	return MoveResult{
		ProjectID:       plan.Project.ID,
		OperationPlanID: operationPlanID,
		Operations:      []ContentOperation{operation},
	}, nil
}

func isDescendantOf(plan model.PlannedProject, itemID, ancestorID uuid.UUID) bool {
	parentByItem := make(map[uuid.UUID]uuid.UUID)
	for _, item := range plan.Items {
		if item.ParentID == nil {
			continue
		}
		parentByItem[item.ID] = *item.ParentID
	}
	for current, ok := parentByItem[itemID]; ok; current, ok = parentByItem[current] {
		if current == ancestorID {
			return true
		}
	}
	return false
}

func visiblePathForItem(plan model.PlannedProject, itemID uuid.UUID) (string, error) {
	for _, object := range plan.StorageObjects {
		if object.ItemID != itemID {
			continue
		}
		switch object.Type {
		case model.StorageObjectTypeFile, model.StorageObjectTypeFolder:
			return object.VisiblePath, nil
		}
	}
	return "", fmt.Errorf("storage object for item %s not found", itemID)
}

func replaceVisiblePathPrefix(path, oldPrefix, newPrefix string) (string, bool) {
	if path == oldPrefix {
		return newPrefix, true
	}
	prefix := oldPrefix + "/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	return newPrefix + strings.TrimPrefix(path, oldPrefix), true
}
