package app

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

func TestServiceInspectAndVerifyShare(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	projectPath := filepath.Join(root, "project"+format.ProjectExtension)
	sharePath := filepath.Join(root, "share"+format.SetExtension)
	projectPassword := "project-password"
	sharePassword := "share-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := project.Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := (project.Executor{OutputRoot: encrypted}).EncryptContent(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := WriteProjectDatabase(ctx, db.Config{
		Path:       projectPath,
		DriverName: db.SQLCipherDriver,
		Password:   projectPassword,
	}, plan); err != nil {
		t.Fatal(err)
	}

	selection := selectShareForTest(t, ctx, projectPath, projectPassword, []string{"source/docs"})
	if err := WriteShareDatabase(ctx, db.Config{
		Path:       sharePath,
		DriverName: db.SQLCipherDriver,
		Password:   sharePassword,
	}, selection.Plan); err != nil {
		t.Fatal(err)
	}
	shareContent := filepath.Join(root, "share-content")
	for _, location := range selection.ContentLocations {
		sourcePath := filepath.Join(encrypted, filepath.FromSlash(location.SourcePath))
		targetPath := filepath.Join(shareContent, filepath.FromSlash(location.TargetPath))
		copyPathForTest(t, sourcePath, targetPath)
	}

	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	shareSummary, err := service.InspectShare(ctx, ShareOpen{
		DatabasePath: sharePath,
		Password:     sharePassword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if shareSummary.DatabaseType != "share" || shareSummary.Files != 1 || shareSummary.TopLevelItems != 1 || !shareSummary.PasswordProtected {
		t.Fatalf("share summary = %+v", shareSummary)
	}

	verify, err := service.VerifyShare(ctx, ShareOpen{
		DatabasePath: sharePath,
		Password:     sharePassword,
	}, shareContent)
	if err != nil {
		t.Fatal(err)
	}
	if verify.Status != "ok" || verify.MissingObjects != 0 || verify.TamperedObjects != 0 || verify.ExtraObjects != 0 {
		t.Fatalf("verify share result = %+v", verify)
	}

	restored := filepath.Join(root, "restored")
	decrypted, err := service.DecryptShare(ctx, DecryptShareInput{
		DatabasePath:  sharePath,
		Password:      sharePassword,
		EncryptedRoot: shareContent,
		OutputRoot:    restored,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}
	if decrypted.DecryptedFiles != 1 || decrypted.RestoredFolders != 1 || decrypted.OutputRoot != restored {
		t.Fatalf("decrypt share result = %+v", decrypted)
	}
	if data, err := os.ReadFile(filepath.Join(restored, "docs", "note.txt")); err != nil || string(data) != "hello" {
		t.Fatalf("restored file data = %q, err = %v", data, err)
	}
}

func TestServiceInspectUnprotectedShareWithoutPassword(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	projectPath := filepath.Join(root, "project"+format.ProjectExtension)
	sharePath := filepath.Join(root, "share"+format.SetExtension)
	projectPassword := "project-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := project.Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteProjectDatabase(ctx, db.Config{
		Path:       projectPath,
		DriverName: db.SQLCipherDriver,
		Password:   projectPassword,
	}, plan); err != nil {
		t.Fatal(err)
	}

	selection := selectShareForTest(t, ctx, projectPath, projectPassword, []string{"source/note.txt"})
	if err := WriteShareDatabase(ctx, db.Config{
		Path:       sharePath,
		DriverName: db.SQLCipherDriver,
		Password:   db.UnprotectedSharePassword,
	}, selection.Plan); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	shareSummary, err := service.InspectShare(ctx, ShareOpen{
		DatabasePath: sharePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if shareSummary.PasswordProtected {
		t.Fatalf("share summary = %+v, want unprotected", shareSummary)
	}
}

func selectShareForTest(t *testing.T, ctx context.Context, projectPath, password string, itemPaths []string) storage.ShareSelection {
	t.Helper()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       projectPath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		t.Fatal(err)
	}
	selection, err := store.SelectShare(ctx, itemPaths, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return selection
}

func copyPathForTest(t *testing.T, source, target string) {
	t.Helper()
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
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
			return copyFileForTest(path, targetPath)
		}); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := copyFileForTest(source, target); err != nil {
		t.Fatal(err)
	}
}

func copyFileForTest(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		_ = output.Close()
		if !committed {
			_ = os.Remove(target)
		}
	}()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	committed = true
	return nil
}
