package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"foldersguard/internal/content"
	"foldersguard/internal/model"
	"foldersguard/internal/progress"
)

type Executor struct {
	OutputRoot string
	Encryptor  content.Encryptor
	AfterFile  func(model.File) error
	// Progress, when set, receives byte-weighted progress for the encryption
	// phase. A nil tracker is safe and ignored.
	Progress *progress.Tracker
	// SkipProgressTotals, when true, leaves the tracker's byte and item totals
	// untouched so a caller can establish a combined total across several
	// EncryptContent calls (for example, applying multiple added items).
	SkipProgressTotals bool
	// Resume, when true, skips a file whose encrypted object(s) already exist
	// instead of re-encrypting it, so an interrupted encryption can continue.
	Resume bool
	// ResumeVerify, when true, additionally authenticates each existing object
	// before skipping it; a present but corrupt object is re-encrypted. It has no
	// effect unless Resume is set.
	ResumeVerify bool
	// ContinueOnError, when true, records a file that fails to encrypt and
	// continues with the remaining files instead of aborting. The default is to
	// abort on the first error.
	ContinueOnError bool
	// OnFileError, when set, is called for each file that fails to encrypt under
	// ContinueOnError, with the file and its error. A failed file's source is
	// never deleted, because AfterFile runs only after a successful encryption.
	OnFileError func(model.File, error)
}

func (e Executor) EncryptContent(ctx context.Context, plan model.PlannedProject) error {
	if e.OutputRoot == "" {
		return fmt.Errorf("output root is required")
	}

	encryptor := e.Encryptor
	encryptor.OutputRoot = e.OutputRoot
	encryptor.OnBytes = e.Progress.AddBytes

	if !e.SkipProgressTotals {
		var totalBytes int64
		for _, file := range plan.Files {
			totalBytes += file.OriginalSize
		}
		e.Progress.SetTotalItems(len(plan.Files))
		e.Progress.SetTotalBytes(totalBytes)
	}

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
		e.Progress.SetItem(filepath.Base(file.SourcePath))

		if e.Resume {
			done, err := e.fileAlreadyEncrypted(ctx, file, visiblePath, partsByFile[file.ID.String()])
			if err != nil {
				return err
			}
			if done {
				// The file is already encrypted. Count its bytes as processed
				// once (the verify read is not fed to progress, to avoid double
				// counting if a partially complete split is re-encrypted), then
				// run AfterFile so source cleanup still applies.
				e.Progress.AddBytes(file.OriginalSize)
				if e.AfterFile != nil {
					if err := e.AfterFile(file); err != nil {
						return fmt.Errorf("post-encrypt file %s: %w", file.ID, err)
					}
				}
				e.Progress.ItemDone()
				continue
			}
		}

		if err := encryptor.EncryptFile(ctx, content.FileSource{
			FileID:       file.ID.String(),
			AbsolutePath: file.SourcePath,
			Key:          file.Key,
			StorageKind:  file.StorageKind,
			VisiblePath:  visiblePath,
			Parts:        partsByFile[file.ID.String()],
		}); err != nil {
			if e.ContinueOnError && ctx.Err() == nil {
				if e.OnFileError != nil {
					e.OnFileError(file, err)
				}
				e.Progress.ItemDone()
				continue
			}
			return fmt.Errorf("encrypt file %s: %w", file.ID, err)
		}
		if e.AfterFile != nil {
			if err := e.AfterFile(file); err != nil {
				return fmt.Errorf("post-encrypt file %s: %w", file.ID, err)
			}
		}
		e.Progress.ItemDone()
	}

	return nil
}

// fileAlreadyEncrypted reports whether a file's encrypted object(s) already
// exist at their visible paths. When ResumeVerify is set it also authenticates
// each object, so a present but corrupt object is treated as incomplete.
func (e Executor) fileAlreadyEncrypted(ctx context.Context, file model.File, visiblePath string, parts []model.Part) (bool, error) {
	check := func(relativePath string, associatedData []byte) (bool, error) {
		absolutePath := filepath.Join(e.OutputRoot, filepath.FromSlash(relativePath))
		if _, err := os.Stat(absolutePath); err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, fmt.Errorf("stat encrypted object %s: %w", relativePath, err)
		}
		if !e.ResumeVerify {
			return true, nil
		}
		if err := content.VerifyObjectFileStream(ctx, file.Key, absolutePath, associatedData, nil); err != nil {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			return false, nil
		}
		return true, nil
	}

	switch file.StorageKind {
	case model.StorageKindSingle:
		return check(visiblePath, []byte("fg-content-v1:file:"+file.ID.String()))
	case model.StorageKindSplit:
		if len(parts) == 0 {
			return false, nil
		}
		for _, part := range parts {
			associatedData := []byte(fmt.Sprintf("fg-content-v1:part:%s:%d:%d:%d", file.ID.String(), part.Index, part.Offset, part.Size))
			ok, err := check(visiblePath+"/"+part.VisibleName.String(), associatedData)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported storage kind %q", file.StorageKind)
	}
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
