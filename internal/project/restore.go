package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"foldersguard/internal/content"
	"foldersguard/internal/fsmeta"
	"foldersguard/internal/model"
	"foldersguard/internal/progress"
)

type Restorer struct {
	EncryptedRoot string
	OutputRoot    string
	NoiseMode     string
	AfterFile     func(RestoredFile) error
	// Progress, when set, receives byte-weighted progress for restore.
	// A nil tracker is safe and ignored.
	Progress *progress.Tracker
}

type RestoredFile struct {
	File                   model.File
	EncryptedPaths         []string
	EncryptedAbsolutePaths []string
}

type RestoreReport struct {
	DecryptedFiles  int
	RestoredFolders int
	SkippedFolders  int
}

func (r Restorer) RestoreContent(ctx context.Context, plan model.PlannedProject) error {
	_, err := r.RestoreContentReport(ctx, plan)
	return err
}

func (r Restorer) RestoreContentReport(ctx context.Context, plan model.PlannedProject) (RestoreReport, error) {
	if r.EncryptedRoot == "" {
		return RestoreReport{}, fmt.Errorf("encrypted root is required")
	}
	if r.OutputRoot == "" {
		return RestoreReport{}, fmt.Errorf("output root is required")
	}

	logicalPaths, err := logicalRealPaths(plan)
	if err != nil {
		return RestoreReport{}, err
	}
	visiblePaths := visiblePathsByItem(plan)
	partsByFile := partsByFileID(plan.Parts)

	itemByID := itemsByID(plan)
	selection, err := r.selectAvailableContent(ctx, plan, itemByID)
	if err != nil {
		return RestoreReport{}, err
	}

	if err := r.createFolders(ctx, plan, logicalPaths, selection.folderIDs); err != nil {
		return RestoreReport{}, err
	}

	report := RestoreReport{
		RestoredFolders: len(selection.folderIDs),
		SkippedFolders:  CountRestorableFolders(plan) - len(selection.folderIDs),
	}

	// Totals reflect only the selected (available) files so progress reaches its
	// total even on a partial restore.
	var totalBytes int64
	var totalItems int
	for _, file := range plan.Files {
		if _, ok := selection.fileIDs[file.ID.String()]; !ok {
			continue
		}
		totalBytes += file.OriginalSize
		totalItems++
	}
	r.Progress.SetTotalItems(totalItems)
	r.Progress.SetTotalBytes(totalBytes)

	for _, file := range plan.Files {
		if _, ok := selection.fileIDs[file.ID.String()]; !ok {
			continue
		}
		if err := ctx.Err(); err != nil {
			return report, err
		}
		var restoredEncryptedPaths []string
		var restoredEncryptedAbsolutePaths []string
		realPath, ok := logicalPaths[file.ID.String()]
		if !ok {
			return report, fmt.Errorf("missing logical path for file %s", file.ID)
		}
		outputPath, err := content.SafeJoin(r.OutputRoot, realPath)
		if err != nil {
			return report, fmt.Errorf("resolve output path for file %s: %w", file.ID, err)
		}
		r.Progress.SetItem(filepath.Base(realPath))

		switch file.StorageKind {
		case model.StorageKindSingle:
			visiblePath, ok := visiblePaths[file.ID.String()]
			if !ok {
				return report, fmt.Errorf("missing visible path for file %s", file.ID)
			}
			sourcePath, ok := selection.sourcePaths[visiblePath]
			if !ok {
				return report, fmt.Errorf("missing selected encrypted path for file %s", file.ID)
			}
			if err := r.restoreSingle(ctx, file, sourcePath, outputPath); err != nil {
				return report, err
			}
			restoredEncryptedPaths = []string{visiblePath}
			restoredEncryptedAbsolutePaths = []string{sourcePath}
		case model.StorageKindSplit:
			visiblePath := visiblePaths[file.ID.String()]
			parts := partsByFile[file.ID.String()]
			sourcePaths, err := selectedPartPaths(visiblePath, parts, selection.sourcePaths)
			if err != nil {
				return report, err
			}
			if err := r.restoreSplit(ctx, file, sourcePaths, parts, outputPath); err != nil {
				return report, err
			}
			restoredEncryptedPaths = encryptedPartPaths(visiblePath, parts)
			restoredEncryptedAbsolutePaths = absolutePartPaths(parts, selection.sourcePaths)
		default:
			return report, fmt.Errorf("unsupported storage kind %q", file.StorageKind)
		}
		item, ok := itemByID[file.ID.String()]
		if !ok {
			return report, fmt.Errorf("missing item for file %s", file.ID)
		}
		if err := fsmeta.Apply(outputPath, metadataFromItem(item)); err != nil {
			return report, fmt.Errorf("restore metadata for file %s: %w", file.ID, err)
		}
		if r.AfterFile != nil {
			restoredFile := RestoredFile{
				File:                   file,
				EncryptedPaths:         restoredEncryptedPaths,
				EncryptedAbsolutePaths: restoredEncryptedAbsolutePaths,
			}
			if err := r.AfterFile(restoredFile); err != nil {
				return report, fmt.Errorf("post-restore file %s: %w", file.ID, err)
			}
		}
		r.Progress.ItemDone()
		report.DecryptedFiles++
	}
	if err := r.restoreFolderMetadata(ctx, plan, logicalPaths, itemByID, selection.folderIDs); err != nil {
		return report, err
	}
	return report, nil
}

func (r Restorer) createFolders(ctx context.Context, plan model.PlannedProject, logicalPaths map[string]string, selected map[string]struct{}) error {
	ids := make([]string, 0, len(logicalPaths))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return pathDepth(logicalPaths[ids[i]]) < pathDepth(logicalPaths[ids[j]])
	})

	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		realPath, ok := logicalPaths[id]
		if !ok {
			return fmt.Errorf("missing logical path for folder %s", id)
		}
		outputPath, err := content.SafeJoin(r.OutputRoot, realPath)
		if err != nil {
			return fmt.Errorf("resolve output path for folder %s: %w", id, err)
		}
		if err := os.MkdirAll(outputPath, 0o755); err != nil {
			return fmt.Errorf("create restored folder %s: %w", realPath, err)
		}
	}
	return nil
}

func (r Restorer) restoreFolderMetadata(ctx context.Context, plan model.PlannedProject, logicalPaths map[string]string, itemByID map[string]model.Item, selected map[string]struct{}) error {
	ids := make([]string, 0, len(logicalPaths))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return pathDepth(logicalPaths[ids[i]]) > pathDepth(logicalPaths[ids[j]])
	})

	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		realPath, ok := logicalPaths[id]
		if !ok {
			return fmt.Errorf("missing logical path for folder %s", id)
		}
		item, ok := itemByID[id]
		if !ok {
			return fmt.Errorf("missing item for folder %s", id)
		}
		outputPath, err := content.SafeJoin(r.OutputRoot, realPath)
		if err != nil {
			return fmt.Errorf("resolve output path for folder %s: %w", id, err)
		}
		if err := fsmeta.Apply(outputPath, metadataFromItem(item)); err != nil {
			return fmt.Errorf("restore metadata for folder %s: %w", id, err)
		}
	}
	return nil
}

func (r Restorer) restoreSingle(ctx context.Context, file model.File, encryptedPath, outputPath string) error {
	ad := []byte("fg-content-v1:file:" + file.ID.String())
	if err := content.OpenObjectFileStream(ctx, file.Key, encryptedPath, outputPath, ad, r.Progress.AddBytes); err != nil {
		return fmt.Errorf("restore file %s: %w", file.ID, err)
	}
	return nil
}

func (r Restorer) restoreSplit(ctx context.Context, file model.File, sourcePaths map[string]string, parts []model.Part, outputPath string) error {
	if len(parts) == 0 {
		return fmt.Errorf("split file %s has no parts", file.ID)
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].Index < parts[j].Index
	})

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create restored split file directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(outputPath), "."+filepath.Base(outputPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary restored split file: %w", err)
	}
	tempPath := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	for _, part := range parts {
		if err := ctx.Err(); err != nil {
			_ = temp.Close()
			return err
		}
		partPath := sourcePaths[part.ID.String()]
		ad := []byte(fmt.Sprintf("fg-content-v1:part:%s:%d:%d:%d", file.ID.String(), part.Index, part.Offset, part.Size))
		input, err := os.Open(partPath)
		if err != nil {
			_ = temp.Close()
			return fmt.Errorf("open encrypted part %s: %w", part.ID, err)
		}
		if err := content.StreamDecrypt(ctx, file.Key, input, temp, ad, r.Progress.AddBytes); err != nil {
			_ = input.Close()
			_ = temp.Close()
			return fmt.Errorf("restore part %s: %w", part.ID, err)
		}
		_ = input.Close()
	}
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restrict restored split file permissions: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close restored split file: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("commit restored split file: %w", err)
	}
	committed = true
	return nil
}

func SafeEncryptedPath(root, visiblePath string) (string, error) {
	return content.SafeJoin(root, visiblePath)
}

func encryptedPartPaths(visiblePath string, parts []model.Part) []string {
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		paths = append(paths, visiblePath+"/"+part.VisibleName.String())
	}
	return paths
}

func logicalRealPaths(plan model.PlannedProject) (map[string]string, error) {
	virtualRoot := isVirtualRoot(plan)
	if !virtualRoot {
		if err := validateRealName(plan.RootItem.RealName); err != nil {
			return nil, fmt.Errorf("invalid root real name: %w", err)
		}
	}
	paths := map[string]string{
		plan.RootItem.ID.String(): plan.RootItem.RealName,
	}
	itemsByParent := make(map[string][]model.Item)
	for _, item := range plan.Items {
		if err := validateRealName(item.RealName); err != nil {
			return nil, fmt.Errorf("invalid real name for item %s: %w", item.ID, err)
		}
		if item.ParentID == nil {
			return nil, fmt.Errorf("non-root item %s has no parent", item.ID)
		}
		itemsByParent[item.ParentID.String()] = append(itemsByParent[item.ParentID.String()], item)
	}

	var walk func(parentID string) error
	walk = func(parentID string) error {
		children := itemsByParent[parentID]
		sort.Slice(children, func(i, j int) bool {
			return children[i].RealName < children[j].RealName
		})
		for _, item := range children {
			parentPath := paths[parentID]
			if parentPath == "" {
				paths[item.ID.String()] = item.RealName
			} else {
				paths[item.ID.String()] = filepath.ToSlash(filepath.Join(parentPath, item.RealName))
			}
			if err := walk(item.ID.String()); err != nil {
				return err
			}
		}
		delete(itemsByParent, parentID)
		return nil
	}
	if err := walk(plan.RootItem.ID.String()); err != nil {
		return nil, err
	}
	if len(itemsByParent) != 0 {
		return nil, fmt.Errorf("items contain missing or cyclic parent references")
	}
	return paths, nil
}

func isVirtualRoot(plan model.PlannedProject) bool {
	return plan.Project.DatabaseType == "share" && plan.RootItem.RealName == ""
}

func itemsByID(plan model.PlannedProject) map[string]model.Item {
	output := map[string]model.Item{
		plan.RootItem.ID.String(): plan.RootItem,
	}
	for _, item := range plan.Items {
		output[item.ID.String()] = item
	}
	return output
}

func metadataFromItem(item model.Item) fsmeta.Metadata {
	return fsmeta.Metadata{
		Mode:              item.OriginalMode,
		ModTime:           item.OriginalModTime,
		AccessTime:        item.OriginalAccessTime,
		BirthTime:         item.OriginalBirthTime,
		WindowsAttributes: item.WindowsAttributes,
		Capabilities:      item.MetadataCaps,
	}
}

func validateRealName(name string) error {
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

func visiblePathsByItem(plan model.PlannedProject) map[string]string {
	paths := make(map[string]string)
	for _, object := range plan.StorageObjects {
		switch object.Type {
		case model.StorageObjectTypeFile, model.StorageObjectTypeFolder:
			paths[object.ItemID.String()] = object.VisiblePath
		}
	}
	return paths
}

func partsByFileID(parts []model.Part) map[string][]model.Part {
	output := make(map[string][]model.Part)
	for _, part := range parts {
		output[part.FileID.String()] = append(output[part.FileID.String()], part)
	}
	return output
}
