package app

import (
	"fmt"
	"os"
	"path/filepath"

	"foldersguard/internal/content"
	"foldersguard/internal/storage"
)

type ContentOperationApplyOptions struct {
	ContentRoot string
	StagingRoot string
}

type AppliedContentOperation struct {
	Type       string
	SourcePath string
	TargetPath string
}

func ValidateStorageContentOperations(operations []storage.ContentOperation, options ContentOperationApplyOptions) error {
	for _, operation := range operations {
		switch operation.Type {
		case "upload":
			if err := validateUploadContent(options.StagingRoot, options.ContentRoot, operation); err != nil {
				return err
			}
		case "move":
			if err := validateMoveContent(options.ContentRoot, operation); err != nil {
				return err
			}
		case "delete":
			if err := validateDeleteContent(options.ContentRoot, operation); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported content operation %q", operation.Type)
		}
	}
	return nil
}

func ApplyStorageContentOperations(operations []storage.ContentOperation, options ContentOperationApplyOptions) ([]AppliedContentOperation, error) {
	applied := make([]AppliedContentOperation, 0, len(operations))
	for _, operation := range operations {
		switch operation.Type {
		case "upload":
			if err := uploadStagedContent(options.StagingRoot, options.ContentRoot, operation); err != nil {
				return applied, err
			}
		case "move":
			if err := moveContent(options.ContentRoot, operation); err != nil {
				return applied, err
			}
		case "delete":
			if err := deleteContent(options.ContentRoot, operation); err != nil {
				return applied, err
			}
		default:
			return applied, fmt.Errorf("unsupported content operation %q", operation.Type)
		}
		applied = append(applied, AppliedContentOperation{
			Type:       operation.Type,
			SourcePath: operation.SourcePath,
			TargetPath: operation.TargetPath,
		})
	}
	return applied, nil
}

func validateUploadContent(stagingRoot, contentRoot string, operation storage.ContentOperation) error {
	source, err := content.SafeJoin(stagingRoot, operation.SourcePath)
	if err != nil {
		return fmt.Errorf("resolve upload source: %w", err)
	}
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("stat upload source: %w", err)
	}
	target, err := content.SafeJoin(contentRoot, operation.TargetPath)
	if err != nil {
		return fmt.Errorf("resolve upload target: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("upload target already exists: %s", operation.TargetPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat upload target: %w", err)
	}
	return nil
}

func validateMoveContent(contentRoot string, operation storage.ContentOperation) error {
	source, err := content.SafeJoin(contentRoot, operation.SourcePath)
	if err != nil {
		return fmt.Errorf("resolve move source: %w", err)
	}
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("stat move source: %w", err)
	}
	target, err := content.SafeJoin(contentRoot, operation.TargetPath)
	if err != nil {
		return fmt.Errorf("resolve move target: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("move target already exists: %s", operation.TargetPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat move target: %w", err)
	}
	return nil
}

func validateDeleteContent(contentRoot string, operation storage.ContentOperation) error {
	target, err := content.SafeJoin(contentRoot, operation.TargetPath)
	if err != nil {
		return fmt.Errorf("resolve delete target: %w", err)
	}
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("stat delete target: %w", err)
	}
	return nil
}

func uploadStagedContent(stagingRoot, contentRoot string, operation storage.ContentOperation) error {
	source, err := content.SafeJoin(stagingRoot, operation.SourcePath)
	if err != nil {
		return fmt.Errorf("resolve upload source: %w", err)
	}
	target, err := content.SafeJoin(contentRoot, operation.TargetPath)
	if err != nil {
		return fmt.Errorf("resolve upload target: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("upload target already exists: %s", operation.TargetPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat upload target: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create upload target parent: %w", err)
	}
	if err := os.Rename(source, target); err == nil {
		return nil
	}
	if err := copyPath(source, target); err != nil {
		return err
	}
	if err := os.RemoveAll(source); err != nil {
		return fmt.Errorf("remove uploaded staging content: %w", err)
	}
	return nil
}

func moveContent(contentRoot string, operation storage.ContentOperation) error {
	source, err := content.SafeJoin(contentRoot, operation.SourcePath)
	if err != nil {
		return fmt.Errorf("resolve move source: %w", err)
	}
	target, err := content.SafeJoin(contentRoot, operation.TargetPath)
	if err != nil {
		return fmt.Errorf("resolve move target: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("move target already exists: %s", operation.TargetPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat move target: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create move target parent: %w", err)
	}
	if err := os.Rename(source, target); err != nil {
		return fmt.Errorf("move encrypted content %s to %s: %w", operation.SourcePath, operation.TargetPath, err)
	}
	return nil
}

func deleteContent(contentRoot string, operation storage.ContentOperation) error {
	target, err := content.SafeJoin(contentRoot, operation.TargetPath)
	if err != nil {
		return fmt.Errorf("resolve delete target: %w", err)
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("delete encrypted content %s: %w", operation.TargetPath, err)
	}
	return nil
}

func copyPath(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat upload source: %w", err)
	}
	if !info.IsDir() {
		return CopyFile(source, target)
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		return CopyFile(path, targetPath)
	})
}
