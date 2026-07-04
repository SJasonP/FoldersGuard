package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
	// Concurrency is the number of files encrypted at once. A value of 1 or less
	// encrypts files sequentially. Concurrency is across files; within-file chunk
	// streaming is unchanged. Folders are always created up front, before any
	// file is encrypted.
	Concurrency int
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

	// cbMu serializes the AfterFile and OnFileError callbacks so a caller can keep
	// simple, non-atomic counters and slices even when files encrypt concurrently.
	var cbMu sync.Mutex

	// encryptOne performs all work for one file. It returns an error only for
	// conditions that must abort the whole operation; under ContinueOnError, a
	// file that fails to encrypt is recorded and nil is returned so the operation
	// proceeds.
	encryptOne := func(ctx context.Context, file model.File) error {
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
					cbMu.Lock()
					err := e.AfterFile(file)
					cbMu.Unlock()
					if err != nil {
						return fmt.Errorf("post-encrypt file %s: %w", file.ID, err)
					}
				}
				e.Progress.ItemDone()
				return nil
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
					cbMu.Lock()
					e.OnFileError(file, err)
					cbMu.Unlock()
				}
				e.Progress.ItemDone()
				return nil
			}
			return fmt.Errorf("encrypt file %s: %w", file.ID, err)
		}
		if e.AfterFile != nil {
			cbMu.Lock()
			err := e.AfterFile(file)
			cbMu.Unlock()
			if err != nil {
				return fmt.Errorf("post-encrypt file %s: %w", file.ID, err)
			}
		}
		e.Progress.ItemDone()
		return nil
	}

	if e.Concurrency > 1 {
		return encryptFilesConcurrent(ctx, plan.Files, e.Concurrency, encryptOne)
	}
	for _, file := range plan.Files {
		if err := encryptOne(ctx, file); err != nil {
			return err
		}
	}
	return nil
}

// encryptFilesConcurrent processes files through a bounded worker pool. The
// first file that returns an error cancels the shared context so the remaining
// workers stop cleanly, and that error is returned. Under ContinueOnError the
// per-file function returns nil for item failures, so the pool runs to
// completion.
func encryptFilesConcurrent(ctx context.Context, files []model.File, workers int, encryptOne func(context.Context, model.File) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan model.File)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				if ctx.Err() != nil {
					continue
				}
				if err := encryptOne(ctx, file); err != nil {
					once.Do(func() {
						firstErr = err
						cancel()
					})
				}
			}
		}()
	}

feed:
	for _, file := range files {
		select {
		case <-ctx.Done():
			break feed
		case jobs <- file:
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}
	return ctx.Err()
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
