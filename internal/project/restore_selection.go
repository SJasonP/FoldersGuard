package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"foldersguard/internal/content"
	"foldersguard/internal/model"
	"foldersguard/internal/noise"
)

type restoreSelection struct {
	fileIDs     map[string]struct{}
	folderIDs   map[string]struct{}
	sourcePaths map[string]string
}

// RestoreSpaceSizes returns per-file output sizes and encrypted source sizes in
// restore order. It uses the same content-selection rules as restoration.
func (r Restorer) RestoreSpaceSizes(ctx context.Context, plan model.PlannedProject) ([]int64, []int64, error) {
	selection, err := r.selectAvailableContent(ctx, plan, itemsByID(plan))
	if err != nil {
		return nil, nil, err
	}
	visiblePaths := visiblePathsByItem(plan)
	partsByFile := partsByFileID(plan.Parts)
	logicalPaths, err := logicalRealPaths(plan)
	if err != nil {
		return nil, nil, err
	}
	var outputs, freed []int64
	for _, file := range plan.Files {
		if _, ok := selection.fileIDs[file.ID.String()]; !ok {
			continue
		}
		if r.Resume {
			outputPath, err := content.SafeJoin(r.OutputRoot, logicalPaths[file.ID.String()])
			if err != nil {
				return nil, nil, err
			}
			if info, err := os.Stat(outputPath); err == nil && !info.IsDir() && info.Size() == file.OriginalSize {
				continue
			} else if err != nil && !os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("stat restored output: %w", err)
			}
		}
		var encryptedSize int64
		switch file.StorageKind {
		case model.StorageKindSingle:
			info, err := os.Stat(selection.sourcePaths[visiblePaths[file.ID.String()]])
			if err != nil {
				return nil, nil, fmt.Errorf("stat encrypted source: %w", err)
			}
			encryptedSize = info.Size()
		case model.StorageKindSplit:
			parts := append([]model.Part(nil), partsByFile[file.ID.String()]...)
			sort.Slice(parts, func(i, j int) bool { return parts[i].Index < parts[j].Index })
			for _, part := range parts {
				path := selection.sourcePaths[visiblePaths[file.ID.String()]+"/"+part.VisibleName.String()]
				info, err := os.Stat(path)
				if err != nil {
					return nil, nil, fmt.Errorf("stat encrypted part: %w", err)
				}
				if r.IncrementalCleanup {
					outputs = append(outputs, part.Size)
					freed = append(freed, info.Size())
					continue
				}
				encryptedSize += info.Size()
			}
		default:
			return nil, nil, fmt.Errorf("unsupported storage kind %q", file.StorageKind)
		}
		if !r.IncrementalCleanup || file.StorageKind != model.StorageKindSplit {
			outputs = append(outputs, file.OriginalSize)
			freed = append(freed, encryptedSize)
		}
	}
	return outputs, freed, nil
}

func (r Restorer) selectAvailableContent(ctx context.Context, plan model.PlannedProject, itemByID map[string]model.Item) (restoreSelection, error) {
	return matchAvailableContent(ctx, r.EncryptedRoot, plan, itemByID, r.NoiseMode)
}

func matchAvailableContent(ctx context.Context, encryptedRoot string, plan model.PlannedProject, itemByID map[string]model.Item, noiseMode string) (restoreSelection, error) {
	objectsByLeaf := make(map[string][]model.StorageObject)
	for _, object := range plan.StorageObjects {
		objectsByLeaf[pathLeaf(object.VisiblePath)] = append(objectsByLeaf[pathLeaf(object.VisiblePath)], object)
	}

	selection := restoreSelection{
		fileIDs:     make(map[string]struct{}),
		folderIDs:   make(map[string]struct{}),
		sourcePaths: make(map[string]string),
	}
	if err := selectRootPath(encryptedRoot, plan, objectsByLeaf, itemByID, &selection); err != nil {
		return restoreSelection{}, err
	}

	err := filepath.WalkDir(encryptedRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == encryptedRoot {
			return nil
		}
		if noise.IgnoreDuringMatching(noiseMode) && noise.IsName(entry.Name()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		selection.addObjectMatches(path, entry.IsDir(), objectsByLeaf, itemByID)
		return nil
	})
	if err != nil {
		return restoreSelection{}, fmt.Errorf("scan encrypted content: %w", err)
	}

	selection.completeSplitFiles(plan)
	selection.addAncestorFolders(plan, itemByID)
	if len(selection.fileIDs) == 0 && len(selection.folderIDs) == 0 {
		return restoreSelection{}, fmt.Errorf("encrypted content path does not contain recognized FG content")
	}
	return selection, nil
}

func selectRootPath(encryptedRoot string, plan model.PlannedProject, objectsByLeaf map[string][]model.StorageObject, itemByID map[string]model.Item, selection *restoreSelection) error {
	info, err := os.Stat(encryptedRoot)
	if err != nil {
		return err
	}
	selection.addObjectMatches(encryptedRoot, info.IsDir(), objectsByLeaf, itemByID)
	if !isVirtualRoot(plan) {
		selection.folderIDs[plan.RootFolder.ID.String()] = struct{}{}
	}
	return nil
}

func (s restoreSelection) addObjectMatches(path string, isDir bool, objectsByLeaf map[string][]model.StorageObject, itemByID map[string]model.Item) {
	for _, object := range objectsByLeaf[filepath.Base(path)] {
		if isDir != (object.Type == model.StorageObjectTypeFolder) {
			continue
		}
		item, ok := itemByID[object.ItemID.String()]
		if !ok {
			continue
		}
		s.sourcePaths[object.VisiblePath] = path
		switch item.Type {
		case model.ItemTypeFile:
			if object.Type == model.StorageObjectTypeFile {
				s.fileIDs[item.ID.String()] = struct{}{}
			}
		case model.ItemTypeFolder:
			if object.Type == model.StorageObjectTypeFolder {
				s.folderIDs[item.ID.String()] = struct{}{}
			}
		}
	}
}

func (s restoreSelection) completeSplitFiles(plan model.PlannedProject) {
	partsByFile := partsByFileID(plan.Parts)
	visiblePaths := visiblePathsByItem(plan)
	for _, file := range plan.Files {
		if file.StorageKind != model.StorageKindSplit {
			continue
		}
		visiblePath := visiblePaths[file.ID.String()]
		if visiblePath == "" {
			continue
		}
		if _, ok := s.sourcePaths[visiblePath]; ok && len(partsByFile[file.ID.String()]) == 0 {
			s.fileIDs[file.ID.String()] = struct{}{}
			continue
		}
		if s.hasAllParts(visiblePath, partsByFile[file.ID.String()]) {
			s.fileIDs[file.ID.String()] = struct{}{}
		}
	}
}

func (s restoreSelection) hasAllParts(visiblePath string, parts []model.Part) bool {
	for _, part := range parts {
		partPath := visiblePath + "/" + part.VisibleName.String()
		if _, ok := s.sourcePaths[partPath]; !ok {
			return false
		}
	}
	return true
}

func (s restoreSelection) addAncestorFolders(plan model.PlannedProject, itemByID map[string]model.Item) {
	for fileID := range s.fileIDs {
		s.addAncestors(fileID, itemByID)
	}
	for folderID := range s.folderIDs {
		s.addAncestors(folderID, itemByID)
	}
	if !isVirtualRoot(plan) && len(s.fileIDs)+len(s.folderIDs) > 0 {
		s.folderIDs[plan.RootFolder.ID.String()] = struct{}{}
	}
	if isVirtualRoot(plan) {
		delete(s.folderIDs, plan.RootFolder.ID.String())
	}
}

func (s restoreSelection) addAncestors(itemID string, itemByID map[string]model.Item) {
	item, ok := itemByID[itemID]
	if !ok {
		return
	}
	for item.ParentID != nil {
		parent, ok := itemByID[item.ParentID.String()]
		if !ok {
			return
		}
		if parent.Type == model.ItemTypeFolder {
			s.folderIDs[parent.ID.String()] = struct{}{}
		}
		item = parent
	}
}

func CountRestorableFolders(plan model.PlannedProject) int {
	if isVirtualRoot(plan) {
		return len(plan.Folders)
	}
	return len(plan.Folders) + 1
}

func selectedPartPaths(visiblePath string, parts []model.Part, sourcePaths map[string]string) (map[string]string, error) {
	output := make(map[string]string, len(parts))
	for _, part := range parts {
		partPath := visiblePath + "/" + part.VisibleName.String()
		sourcePath, ok := sourcePaths[partPath]
		if ok {
			output[part.ID.String()] = sourcePath
		}
		if output[part.ID.String()] == "" {
			return nil, fmt.Errorf("missing selected encrypted part %s", part.ID)
		}
	}
	return output, nil
}

func absolutePartPaths(parts []model.Part, sourcePaths map[string]string) []string {
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		for visiblePath, sourcePath := range sourcePaths {
			if pathLeaf(visiblePath) == part.VisibleName.String() {
				paths = append(paths, sourcePath)
				break
			}
		}
	}
	return paths
}

func pathLeaf(path string) string {
	return filepath.Base(filepath.FromSlash(path))
}
