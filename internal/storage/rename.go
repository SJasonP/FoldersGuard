package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/model"
)

type RenameResult struct {
	ProjectID uuid.UUID
	ItemID    uuid.UUID
	OldName   string
	NewName   string
}

func (s *Store) RenameItem(ctx context.Context, itemPath, newName string, now time.Time) (RenameResult, error) {
	if err := validateRealNameSegment(newName); err != nil {
		return RenameResult{}, fmt.Errorf("invalid new name: %w", err)
	}
	plan, err := s.ReadPlannedProject(ctx)
	if err != nil {
		return RenameResult{}, err
	}
	item, err := itemByRealPath(plan, itemPath)
	if err != nil {
		return RenameResult{}, err
	}
	if item.ParentID == nil {
		return RenameResult{}, fmt.Errorf("root item cannot be renamed")
	}
	if siblingNameExists(plan, *item.ParentID, item.ID, newName) {
		return RenameResult{}, fmt.Errorf("sibling name %q already exists", newName)
	}

	updatedAt := formatTime(now.UTC())
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RenameResult{}, fmt.Errorf("begin rename item: %w", err)
	}
	defer rollback(tx)

	result, err := tx.ExecContext(ctx, `
UPDATE items
SET real_name = ?, sort_name = ?, updated_at = ?
WHERE item_id = ?`,
		newName,
		sortName(newName),
		updatedAt,
		item.ID.String(),
	)
	if err != nil {
		return RenameResult{}, fmt.Errorf("rename item %s: %w", item.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return RenameResult{}, fmt.Errorf("read rename row count: %w", err)
	}
	if affected != 1 {
		return RenameResult{}, fmt.Errorf("rename affected %d rows, want 1", affected)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE meta SET value = ? WHERE key = 'updated_at'`, updatedAt); err != nil {
		return RenameResult{}, fmt.Errorf("update metadata timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return RenameResult{}, fmt.Errorf("commit rename item: %w", err)
	}
	return RenameResult{
		ProjectID: plan.Project.ID,
		ItemID:    item.ID,
		OldName:   item.RealName,
		NewName:   newName,
	}, nil
}

func itemByRealPath(plan model.PlannedProject, itemPath string) (model.Item, error) {
	cleanPath := filepath.ToSlash(filepath.Clean(itemPath))
	if cleanPath == "." || cleanPath == "" {
		return model.Item{}, fmt.Errorf("item path is required")
	}

	paths := map[string]model.Item{
		plan.RootItem.RealName: plan.RootItem,
	}
	itemsByParent := make(map[string][]model.Item)
	for _, item := range plan.Items {
		if item.ParentID == nil {
			return model.Item{}, fmt.Errorf("item %s has no parent", item.ID)
		}
		itemsByParent[item.ParentID.String()] = append(itemsByParent[item.ParentID.String()], item)
	}

	var walk func(parentID string, parentPath string)
	walk = func(parentID string, parentPath string) {
		children := itemsByParent[parentID]
		delete(itemsByParent, parentID)
		for _, child := range children {
			childPath := filepath.ToSlash(filepath.Join(parentPath, child.RealName))
			paths[childPath] = child
			walk(child.ID.String(), childPath)
		}
	}
	walk(plan.RootItem.ID.String(), plan.RootItem.RealName)
	if len(itemsByParent) != 0 {
		return model.Item{}, fmt.Errorf("items contain missing or cyclic parent references")
	}

	item, ok := paths[cleanPath]
	if !ok {
		return model.Item{}, fmt.Errorf("item path %q not found", itemPath)
	}
	return item, nil
}

func siblingNameExists(plan model.PlannedProject, parentID uuid.UUID, itemID uuid.UUID, newName string) bool {
	for _, item := range plan.Items {
		if item.ID == itemID || item.ParentID == nil || *item.ParentID != parentID {
			continue
		}
		if item.RealName == newName {
			return true
		}
	}
	if plan.RootItem.ID != itemID && plan.RootItem.ParentID != nil && *plan.RootItem.ParentID == parentID && plan.RootItem.RealName == newName {
		return true
	}
	return false
}

func validateRealNameSegment(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("absolute name rejected")
	}
	clean := filepath.Clean(name)
	if clean != name || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("path-like name rejected")
	}
	return nil
}
