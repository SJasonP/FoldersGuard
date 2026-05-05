package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"foldersguard/internal/content"
	"foldersguard/internal/model"
)

type VerifyReport struct {
	CheckedObjects  int
	MissingObjects  int
	TamperedObjects int
	ExtraObjects    int
}

func (r VerifyReport) OK() bool {
	return r.MissingObjects == 0 && r.TamperedObjects == 0 && r.ExtraObjects == 0
}

type Verifier struct {
	EncryptedRoot string
}

func (v Verifier) VerifyContent(ctx context.Context, plan model.PlannedProject) (VerifyReport, error) {
	if v.EncryptedRoot == "" {
		return VerifyReport{}, fmt.Errorf("encrypted root is required")
	}

	report := VerifyReport{}
	expected := make(map[string]struct{})
	visiblePaths := visiblePathsByItem(plan)
	partsByFile := partsByFileID(plan.Parts)

	for _, object := range plan.StorageObjects {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		if object.Type != model.StorageObjectTypeFolder {
			continue
		}
		if err := v.verifyFolder(object.VisiblePath, expected, &report); err != nil {
			return report, err
		}
	}

	for _, file := range plan.Files {
		if err := ctx.Err(); err != nil {
			return report, err
		}

		switch file.StorageKind {
		case model.StorageKindSingle:
			visiblePath, ok := visiblePaths[file.ID.String()]
			if !ok {
				return report, fmt.Errorf("missing visible path for file %s", file.ID)
			}
			if err := v.verifyObject(ctx, file.Key, visiblePath, []byte("fg-content-v1:file:"+file.ID.String()), expected, &report); err != nil {
				return report, err
			}
		case model.StorageKindSplit:
			visiblePath, ok := visiblePaths[file.ID.String()]
			if !ok {
				return report, fmt.Errorf("missing visible path for split file %s", file.ID)
			}
			for _, part := range partsByFile[file.ID.String()] {
				partPath := visiblePath + "/" + part.VisibleName.String()
				ad := []byte(fmt.Sprintf("fg-content-v1:part:%s:%d:%d:%d", file.ID.String(), part.Index, part.Offset, part.Size))
				if err := v.verifyObject(ctx, file.Key, partPath, ad, expected, &report); err != nil {
					return report, err
				}
			}
		default:
			return report, fmt.Errorf("unsupported storage kind %q", file.StorageKind)
		}
	}

	extra, err := countExtraObjects(ctx, v.EncryptedRoot, expected)
	if err != nil {
		return report, err
	}
	report.ExtraObjects = extra
	return report, nil
}

func (v Verifier) verifyFolder(visiblePath string, expected map[string]struct{}, report *VerifyReport) error {
	cleanPath := filepath.Clean(filepath.FromSlash(visiblePath))
	expected[cleanPath] = struct{}{}
	report.CheckedObjects++

	folderPath, err := content.SafeJoin(v.EncryptedRoot, visiblePath)
	if err != nil {
		return fmt.Errorf("resolve encrypted folder %s: %w", visiblePath, err)
	}
	info, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			report.MissingObjects++
			return nil
		}
		report.TamperedObjects++
		return nil
	}
	if !info.IsDir() {
		report.TamperedObjects++
	}
	return nil
}

func (v Verifier) verifyObject(ctx context.Context, key []byte, visiblePath string, associatedData []byte, expected map[string]struct{}, report *VerifyReport) error {
	expected[filepath.Clean(filepath.FromSlash(visiblePath))] = struct{}{}
	report.CheckedObjects++

	encryptedPath, err := content.SafeJoin(v.EncryptedRoot, visiblePath)
	if err != nil {
		return fmt.Errorf("resolve encrypted object %s: %w", visiblePath, err)
	}
	if _, err := os.Stat(encryptedPath); err != nil {
		if os.IsNotExist(err) {
			report.MissingObjects++
			return nil
		}
		report.TamperedObjects++
		return nil
	}
	if _, err := content.OpenObjectFromFile(ctx, key, encryptedPath, associatedData); err != nil {
		report.TamperedObjects++
	}
	return nil
}

func countExtraObjects(ctx context.Context, root string, expected map[string]struct{}) (int, error) {
	extra := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if _, ok := expected[filepath.Clean(relative)]; !ok {
			extra++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walk encrypted content: %w", err)
	}
	return extra, nil
}
