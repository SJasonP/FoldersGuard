package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

type ContentOperation struct {
	Type       string
	SourcePath string
	TargetPath string
}

type RemoveResult struct {
	ProjectID       uuid.UUID
	OperationPlanID uuid.UUID
	Operations      []ContentOperation
}

func (s *Store) PlanRemove(ctx context.Context, itemPath string) (uuid.UUID, []ContentOperation, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}
	item, err := itemByRealPath(plan, itemPath)
	if err != nil {
		return uuid.Nil, nil, err
	}
	if item.ParentID == nil {
		return uuid.Nil, nil, fmt.Errorf("root item cannot be removed")
	}
	operation, err := deleteOperationForItem(plan, item.ID)
	if err != nil {
		return uuid.Nil, nil, err
	}
	return plan.Project.ID, []ContentOperation{operation}, nil
}

func (s *Store) RemoveItem(ctx context.Context, itemPath string, now time.Time) (RemoveResult, error) {
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return RemoveResult{}, err
	}
	item, err := itemByRealPath(plan, itemPath)
	if err != nil {
		return RemoveResult{}, err
	}
	if item.ParentID == nil {
		return RemoveResult{}, fmt.Errorf("root item cannot be removed")
	}

	removedSize, err := itemOriginalSize(plan, item)
	if err != nil {
		return RemoveResult{}, err
	}
	itemIDs := subtreeItemIDs(plan, item.ID)
	operation, err := deleteOperationForItem(plan, item.ID)
	if err != nil {
		return RemoveResult{}, err
	}

	updatedAt := formatTime(now.UTC())
	operationPlanID := uuid.New()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RemoveResult{}, fmt.Errorf("begin remove item: %w", err)
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
		return RemoveResult{}, fmt.Errorf("insert remove operation plan: %w", err)
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
		nil,
		operation.TargetPath,
		nil,
	); err != nil {
		return RemoveResult{}, fmt.Errorf("insert remove operation step: %w", err)
	}

	for _, id := range itemIDs {
		idText := id.String()
		if _, err := tx.ExecContext(ctx, `DELETE FROM storage_objects WHERE item_id = ?`, idText); err != nil {
			return RemoveResult{}, fmt.Errorf("delete storage objects for %s: %w", id, err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM parts WHERE file_id = ?`, idText); err != nil {
			return RemoveResult{}, fmt.Errorf("delete parts for %s: %w", id, err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE file_id = ?`, idText); err != nil {
			return RemoveResult{}, fmt.Errorf("delete file %s: %w", id, err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM folders WHERE folder_id = ?`, idText); err != nil {
			return RemoveResult{}, fmt.Errorf("delete folder %s: %w", id, err)
		}
	}
	for i := len(itemIDs) - 1; i >= 0; i-- {
		id := itemIDs[i]
		if _, err := tx.ExecContext(ctx, `DELETE FROM items WHERE item_id = ?`, id.String()); err != nil {
			return RemoveResult{}, fmt.Errorf("delete item %s: %w", id, err)
		}
	}
	if err := updateFolderSizeAncestors(ctx, tx, *item.ParentID, -removedSize); err != nil {
		return RemoveResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value = ? WHERE key = 'updated_at'`, updatedAt); err != nil {
		return RemoveResult{}, fmt.Errorf("update metadata timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return RemoveResult{}, fmt.Errorf("commit remove item: %w", err)
	}

	return RemoveResult{
		ProjectID:       plan.Project.ID,
		OperationPlanID: operationPlanID,
		Operations:      []ContentOperation{operation},
	}, nil
}

func subtreeItemIDs(plan model.PlannedProject, rootID uuid.UUID) []uuid.UUID {
	children := make(map[uuid.UUID][]uuid.UUID)
	for _, item := range plan.Items {
		if item.ParentID == nil {
			continue
		}
		children[*item.ParentID] = append(children[*item.ParentID], item.ID)
	}

	var ids []uuid.UUID
	var walk func(uuid.UUID)
	walk = func(id uuid.UUID) {
		ids = append(ids, id)
		for _, childID := range children[id] {
			walk(childID)
		}
	}
	walk(rootID)
	return ids
}

func deleteOperationForItem(plan model.PlannedProject, itemID uuid.UUID) (ContentOperation, error) {
	for _, object := range plan.StorageObjects {
		if object.ItemID != itemID {
			continue
		}
		switch object.Type {
		case model.StorageObjectTypeFile, model.StorageObjectTypeFolder:
			return ContentOperation{
				Type:       "delete",
				TargetPath: object.VisiblePath,
			}, nil
		}
	}
	return ContentOperation{}, fmt.Errorf("storage object for item %s not found", itemID)
}
