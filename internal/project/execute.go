package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"foldersguard/internal/content"
	"foldersguard/internal/model"
)

type Executor struct {
	OutputRoot string
	Encryptor  content.Encryptor
	AfterFile  func(model.File) error
}

func (e Executor) EncryptContent(ctx context.Context, plan model.PlannedProject) error {
	if e.OutputRoot == "" {
		return fmt.Errorf("output root is required")
	}

	encryptor := e.Encryptor
	encryptor.OutputRoot = e.OutputRoot

	if err := e.createFolders(ctx, plan); err != nil {
		return err
	}

	partsByFile := make(map[string][]model.Part)
	for _, part := range plan.Parts {
		partsByFile[part.FileID.String()] = append(partsByFile[part.FileID.String()], part)
	}
	visiblePathByItem := make(map[string]string)
	for _, object := range plan.StorageObjects {
		switch object.Type {
		case model.StorageObjectTypeFile, model.StorageObjectTypeFolder:
			visiblePathByItem[object.ItemID.String()] = object.VisiblePath
		}
	}

	for _, file := range plan.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		visiblePath, ok := visiblePathByItem[file.ID.String()]
		if !ok {
			return fmt.Errorf("missing visible path for file %s", file.ID)
		}
		if err := encryptor.EncryptFile(ctx, content.FileSource{
			FileID:       file.ID.String(),
			AbsolutePath: file.SourcePath,
			Key:          file.Key,
			StorageKind:  file.StorageKind,
			VisiblePath:  visiblePath,
			Parts:        partsByFile[file.ID.String()],
		}); err != nil {
			return fmt.Errorf("encrypt file %s: %w", file.ID, err)
		}
		if e.AfterFile != nil {
			if err := e.AfterFile(file); err != nil {
				return fmt.Errorf("post-encrypt file %s: %w", file.ID, err)
			}
		}
	}

	return nil
}

func (e Executor) createFolders(ctx context.Context, plan model.PlannedProject) error {
	for _, object := range plan.StorageObjects {
		if err := ctx.Err(); err != nil {
			return err
		}
		if object.Type != model.StorageObjectTypeFolder {
			continue
		}
		path := filepath.Join(e.OutputRoot, filepath.FromSlash(object.VisiblePath))
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create output folder %s: %w", object.VisiblePath, err)
		}
	}
	return nil
}
