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
)

type Restorer struct {
	EncryptedRoot string
	OutputRoot    string
	AfterFile     func(RestoredFile) error
}

type RestoredFile struct {
	File           model.File
	EncryptedPaths []string
}

func (r Restorer) RestoreContent(ctx context.Context, plan model.PlannedProject) error {
	if r.EncryptedRoot == "" {
		return fmt.Errorf("encrypted root is required")
	}
	if r.OutputRoot == "" {
		return fmt.Errorf("output root is required")
	}

	logicalPaths, err := logicalRealPaths(plan)
	if err != nil {
		return err
	}
	visiblePaths := visiblePathsByItem(plan)
	partsByFile := partsByFileID(plan.Parts)

	itemByID := itemsByID(plan)

	if err := r.createFolders(ctx, plan, logicalPaths); err != nil {
		return err
	}

	for _, file := range plan.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		var restoredEncryptedPaths []string
		realPath, ok := logicalPaths[file.ID.String()]
		if !ok {
			return fmt.Errorf("missing logical path for file %s", file.ID)
		}
		outputPath, err := content.SafeJoin(r.OutputRoot, realPath)
		if err != nil {
			return fmt.Errorf("resolve output path for file %s: %w", file.ID, err)
		}

		switch file.StorageKind {
		case model.StorageKindSingle:
			visiblePath, ok := visiblePaths[file.ID.String()]
			if !ok {
				return fmt.Errorf("missing visible path for file %s", file.ID)
			}
			if err := r.restoreSingle(ctx, file, visiblePath, outputPath); err != nil {
				return err
			}
			restoredEncryptedPaths = []string{visiblePath}
		case model.StorageKindSplit:
			visiblePath := visiblePaths[file.ID.String()]
			parts := partsByFile[file.ID.String()]
			if err := r.restoreSplit(ctx, file, visiblePath, parts, outputPath); err != nil {
				return err
			}
			restoredEncryptedPaths = encryptedPartPaths(visiblePath, parts)
		default:
			return fmt.Errorf("unsupported storage kind %q", file.StorageKind)
		}
		item, ok := itemByID[file.ID.String()]
		if !ok {
			return fmt.Errorf("missing item for file %s", file.ID)
		}
		if err := fsmeta.Apply(outputPath, metadataFromItem(item)); err != nil {
			return fmt.Errorf("restore metadata for file %s: %w", file.ID, err)
		}
		if r.AfterFile != nil {
			if err := r.AfterFile(RestoredFile{File: file, EncryptedPaths: restoredEncryptedPaths}); err != nil {
				return fmt.Errorf("post-restore file %s: %w", file.ID, err)
			}
		}
	}
	if err := r.restoreFolderMetadata(ctx, plan, logicalPaths, itemByID); err != nil {
		return err
	}
	return nil
}

func (r Restorer) createFolders(ctx context.Context, plan model.PlannedProject, logicalPaths map[string]string) error {
	ids := make([]string, 0, len(logicalPaths))
	folderIDs := make(map[string]struct{})
	if !isVirtualRoot(plan) {
		folderIDs[plan.RootFolder.ID.String()] = struct{}{}
	}
	for _, folder := range plan.Folders {
		folderIDs[folder.ID.String()] = struct{}{}
	}
	for id := range folderIDs {
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

func (r Restorer) restoreFolderMetadata(ctx context.Context, plan model.PlannedProject, logicalPaths map[string]string, itemByID map[string]model.Item) error {
	ids := make([]string, 0, len(logicalPaths))
	folderIDs := make(map[string]struct{})
	if !isVirtualRoot(plan) {
		folderIDs[plan.RootFolder.ID.String()] = struct{}{}
	}
	for _, folder := range plan.Folders {
		folderIDs[folder.ID.String()] = struct{}{}
	}
	for id := range folderIDs {
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

func (r Restorer) restoreSingle(ctx context.Context, file model.File, visiblePath, outputPath string) error {
	encryptedPath, err := SafeEncryptedPath(r.EncryptedRoot, visiblePath)
	if err != nil {
		return fmt.Errorf("resolve encrypted file %s: %w", file.ID, err)
	}
	ad := []byte("fg-content-v1:file:" + file.ID.String())
	if err := content.OpenObjectFile(ctx, file.Key, encryptedPath, outputPath, ad); err != nil {
		return fmt.Errorf("restore file %s: %w", file.ID, err)
	}
	return nil
}

func (r Restorer) restoreSplit(ctx context.Context, file model.File, visiblePath string, parts []model.Part, outputPath string) error {
	if visiblePath == "" {
		return fmt.Errorf("missing visible path for split file %s", file.ID)
	}
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
		partPath, err := SafeEncryptedPath(r.EncryptedRoot, visiblePath+"/"+part.VisibleName.String())
		if err != nil {
			_ = temp.Close()
			return fmt.Errorf("resolve encrypted part %s: %w", part.ID, err)
		}
		ad := []byte(fmt.Sprintf("fg-content-v1:part:%s:%d:%d:%d", file.ID.String(), part.Index, part.Offset, part.Size))
		partPlaintext, err := content.OpenObjectFromFile(ctx, file.Key, partPath, ad)
		if err != nil {
			_ = temp.Close()
			return fmt.Errorf("restore part %s: %w", part.ID, err)
		}
		if _, err := temp.Write(partPlaintext); err != nil {
			_ = temp.Close()
			return fmt.Errorf("write restored part %s: %w", part.ID, err)
		}
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
