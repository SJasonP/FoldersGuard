package project

import (
	"context"

	"foldersguard/internal/model"
	"foldersguard/internal/noise"
)

type ContentMatch struct {
	fileIDs   map[string]struct{}
	folderIDs map[string]struct{}
}

func MatchAvailableContent(ctx context.Context, encryptedRoot string, plan model.PlannedProject) (ContentMatch, error) {
	return MatchAvailableContentWithNoiseMode(ctx, encryptedRoot, plan, noise.ModeDoNotIgnore)
}

func MatchAvailableContentWithNoiseMode(ctx context.Context, encryptedRoot string, plan model.PlannedProject, noiseMode string) (ContentMatch, error) {
	itemByID := itemsByID(plan)
	selection, err := matchAvailableContent(ctx, encryptedRoot, plan, itemByID, noiseMode)
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
