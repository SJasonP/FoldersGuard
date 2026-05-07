package project

import (
	"context"

	"foldersguard/internal/model"
)

type ContentMatch struct {
	fileIDs   map[string]struct{}
	folderIDs map[string]struct{}
}

func MatchAvailableContent(ctx context.Context, encryptedRoot string, plan model.PlannedProject) (ContentMatch, error) {
	itemByID := itemsByID(plan)
	selection, err := matchAvailableContent(ctx, encryptedRoot, plan, itemByID)
	if err != nil {
		return ContentMatch{}, err
	}
	return ContentMatch{
		fileIDs:   selection.fileIDs,
		folderIDs: selection.folderIDs,
	}, nil
}

func (m ContentMatch) HasFile(id string) bool {
	_, ok := m.fileIDs[id]
	return ok
}

func (m ContentMatch) HasFolder(id string) bool {
	_, ok := m.folderIDs[id]
	return ok
}
