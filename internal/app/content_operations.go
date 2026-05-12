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

func ApplyStorageContentOperationsWithCommit(operations []storage.ContentOperation, options ContentOperationApplyOptions, commit func() error) ([]AppliedContentOperation, error) {
	if commit == nil {
		return nil, fmt.Errorf("commit function is required")
	}
	if err := ValidateStorageContentOperations(operations, options); err != nil {
		return nil, err
	}

	transaction := contentOperationTransaction{
		options: options,
	}
	applied, err := transaction.apply(operations)
	if err != nil {
		rollbackErr := transaction.rollback()
		transaction.cleanup()
		if rollbackErr != nil {
			return applied, fmt.Errorf("%w; rollback failed: %v", err, rollbackErr)
		}
		return applied, err
	}
	if err := commit(); err != nil {
		rollbackErr := transaction.rollback()
		transaction.cleanup()
		if rollbackErr != nil {
			return applied, fmt.Errorf("%w; rollback failed: %v", err, rollbackErr)
		}
		return applied, err
	}
	if err := transaction.cleanup(); err != nil {
		return applied, err
	}
	return applied, nil
}

type contentOperationTransaction struct {
	options    ContentOperationApplyOptions
	deleteRoot string
	steps      []contentOperationStep
}

type contentOperationStep struct {
	applied  AppliedContentOperation
	rollback func() error
	cleanup  func() error
}

func (t *contentOperationTransaction) apply(operations []storage.ContentOperation) ([]AppliedContentOperation, error) {
	applied := make([]AppliedContentOperation, 0, len(operations))
	for _, operation := range operations {
		step, err := t.applyOne(operation)
		if err != nil {
			return applied, err
		}
		t.steps = append(t.steps, step)
		applied = append(applied, step.applied)
	}
	return applied, nil
}

func (t *contentOperationTransaction) applyOne(operation storage.ContentOperation) (contentOperationStep, error) {
	switch operation.Type {
	case "upload":
		return t.upload(operation)
	case "move":
		return t.move(operation)
	case "delete":
		return t.delete(operation)
	default:
		return contentOperationStep{}, fmt.Errorf("unsupported content operation %q", operation.Type)
	}
}

func (t *contentOperationTransaction) upload(operation storage.ContentOperation) (contentOperationStep, error) {
	if err := uploadStagedContent(t.options.StagingRoot, t.options.ContentRoot, operation); err != nil {
		return contentOperationStep{}, err
	}
	target, err := content.SafeJoin(t.options.ContentRoot, operation.TargetPath)
	if err != nil {
		return contentOperationStep{}, fmt.Errorf("resolve upload rollback target: %w", err)
	}
	return contentOperationStep{
		applied: appliedContentOperation(operation),
		rollback: func() error {
			if err := os.RemoveAll(target); err != nil {
				return fmt.Errorf("rollback uploaded content %s: %w", operation.TargetPath, err)
			}
			return nil
		},
	}, nil
}

func (t *contentOperationTransaction) move(operation storage.ContentOperation) (contentOperationStep, error) {
	if err := moveContent(t.options.ContentRoot, operation); err != nil {
		return contentOperationStep{}, err
	}
	source, err := content.SafeJoin(t.options.ContentRoot, operation.SourcePath)
	if err != nil {
		return contentOperationStep{}, fmt.Errorf("resolve move rollback source: %w", err)
	}
	target, err := content.SafeJoin(t.options.ContentRoot, operation.TargetPath)
	if err != nil {
		return contentOperationStep{}, fmt.Errorf("resolve move rollback target: %w", err)
	}
	return contentOperationStep{
		applied: appliedContentOperation(operation),
		rollback: func() error {
			if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
				return fmt.Errorf("create move rollback parent: %w", err)
			}
			if err := os.Rename(target, source); err != nil {
				return fmt.Errorf("rollback moved content %s to %s: %w", operation.TargetPath, operation.SourcePath, err)
			}
			return nil
		},
	}, nil
}

func (t *contentOperationTransaction) delete(operation storage.ContentOperation) (contentOperationStep, error) {
	target, err := content.SafeJoin(t.options.ContentRoot, operation.TargetPath)
	if err != nil {
		return contentOperationStep{}, fmt.Errorf("resolve delete target: %w", err)
	}
	if t.deleteRoot == "" {
		root, err := os.MkdirTemp(filepath.Dir(t.options.ContentRoot), ".fg-delete-rollback-*")
		if err != nil {
			return contentOperationStep{}, fmt.Errorf("create delete rollback directory: %w", err)
		}
		t.deleteRoot = root
	}
	backup := filepath.Join(t.deleteRoot, fmt.Sprintf("%06d", len(t.steps)))
	if err := os.Rename(target, backup); err != nil {
		return contentOperationStep{}, fmt.Errorf("stage deleted content %s: %w", operation.TargetPath, err)
	}
	return contentOperationStep{
		applied: appliedContentOperation(operation),
		rollback: func() error {
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("create delete rollback parent: %w", err)
			}
			if err := os.Rename(backup, target); err != nil {
				return fmt.Errorf("rollback deleted content %s: %w", operation.TargetPath, err)
			}
			return nil
		},
		cleanup: func() error {
			if err := os.RemoveAll(backup); err != nil {
				return fmt.Errorf("remove staged deleted content %s: %w", operation.TargetPath, err)
			}
			return nil
		},
	}, nil
}

func (t *contentOperationTransaction) rollback() error {
	for i := len(t.steps) - 1; i >= 0; i-- {
		if t.steps[i].rollback == nil {
			continue
		}
		if err := t.steps[i].rollback(); err != nil {
			return err
		}
	}
	return nil
}

func (t *contentOperationTransaction) cleanup() error {
	for _, step := range t.steps {
		if step.cleanup == nil {
			continue
		}
		if err := step.cleanup(); err != nil {
			return err
		}
	}
	if t.deleteRoot != "" {
		if err := os.RemoveAll(t.deleteRoot); err != nil {
			return fmt.Errorf("remove delete rollback directory: %w", err)
		}
		t.deleteRoot = ""
	}
	return nil
}

func appliedContentOperation(operation storage.ContentOperation) AppliedContentOperation {
	return AppliedContentOperation{
		Type:       operation.Type,
		SourcePath: operation.SourcePath,
		TargetPath: operation.TargetPath,
	}
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
