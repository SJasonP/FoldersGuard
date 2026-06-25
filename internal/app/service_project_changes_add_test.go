package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceApplyProjectAddWritesDesktopStagedContentAndManualGuideResult(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(root, "source")
	addSource := filepath.Join(root, "new.txt")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.Mkdir(filepath.Join(home, "Desktop"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "old.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(addSource, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID: created.ProjectID,
		Password:  password,
		AddChanges: []ProjectAddChange{{
			SourcePath:       addSource,
			TargetFolderPath: "source",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedAdds != 1 || len(result.ContentOperations) != 1 || !result.ManualContentGuide || result.StagedContentPath == "" {
		t.Fatalf("apply result = %+v", result)
	}
	if !result.StagedContentOnDesktop || result.StagedContentName == "" || filepath.Dir(result.StagedContentPath) != filepath.Join(home, "Desktop") {
		t.Fatalf("staged desktop result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/new.txt") {
		t.Fatalf("browser state missing added file: %+v", result.BrowserState.Items)
	}
	assertExists(t, filepath.Join(result.StagedContentPath, filepath.FromSlash(result.ContentOperations[0].SourcePath)))
}

func TestServiceApplyProjectAddUploadsConnectedContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	addSource := filepath.Join(root, "new.txt")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "old.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(addSource, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		AddChanges: []ProjectAddChange{{
			SourcePath:       addSource,
			TargetFolderPath: "source",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedAdds != 1 || len(result.AppliedContentChanges) != 1 || result.ManualContentGuide || result.StagedContentPath != "" {
		t.Fatalf("apply result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/new.txt") {
		t.Fatalf("browser state missing added file: %+v", result.BrowserState.Items)
	}
	assertExists(t, filepath.Join(encrypted, filepath.FromSlash(result.AppliedContentChanges[0].TargetPath)))
	if err := os.WriteFile(filepath.Join(encrypted, ".DS_Store"), []byte("finder metadata"), 0o600); err != nil {
		t.Fatal(err)
	}

	verify, err := service.Verify(ctx, DatabaseOpen{
		ProjectRef: created.ProjectID,
		Password:   password,
	}, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if verify.Status != "ok" || verify.MissingObjects != 0 || verify.TamperedObjects != 0 || verify.ExtraObjects != 0 {
		t.Fatalf("verify after connected add = %+v", verify)
	}
}

// addTestProject creates a project (keeping its source) and returns the service
// and project id, ready for add-change tests.
func addTestProject(t *testing.T, dataDir, source, encrypted, password string) (Service, string) {
	t.Helper()
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "old.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(context.Background(), CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}
	return service, created.ProjectID
}

func TestServiceApplyProjectAddDeletesSourceWhenCleanupDelete(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	addSource := filepath.Join(root, "new.txt")
	if err := os.WriteFile(addSource, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	encrypted := filepath.Join(root, "encrypted")
	service, projectID := addTestProject(t, filepath.Join(root, "data"), filepath.Join(root, "source"), encrypted, "project-password")

	// Default settings use delete source cleanup.
	if _, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID:     projectID,
		Password:      "project-password",
		EncryptedRoot: encrypted,
		AddChanges:    []ProjectAddChange{{SourcePath: addSource, TargetFolderPath: "source"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(addSource); !os.IsNotExist(err) {
		t.Fatalf("expected added source file to be deleted, stat err = %v", err)
	}
}

func TestServiceApplyProjectAddKeepsSourceWhenCleanupKeep(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	addSource := filepath.Join(root, "new.txt")
	if err := os.WriteFile(addSource, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	encrypted := filepath.Join(root, "encrypted")
	service, projectID := addTestProject(t, filepath.Join(root, "data"), filepath.Join(root, "source"), encrypted, "project-password")

	if _, err := service.SaveSettings(Settings{SourceCleanupMode: SourceCleanupKeep}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID:     projectID,
		Password:      "project-password",
		EncryptedRoot: encrypted,
		AddChanges:    []ProjectAddChange{{SourcePath: addSource, TargetFolderPath: "source"}},
	}); err != nil {
		t.Fatal(err)
	}
	assertExists(t, addSource)
}

func TestServiceApplyProjectAddDeletesSourceDirectory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	addDir := filepath.Join(root, "added-dir")
	if err := os.MkdirAll(filepath.Join(addDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(addDir, "a.txt"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(addDir, "nested", "b.txt"), []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}
	encrypted := filepath.Join(root, "encrypted")
	service, projectID := addTestProject(t, filepath.Join(root, "data"), filepath.Join(root, "source"), encrypted, "project-password")

	if _, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID:     projectID,
		Password:      "project-password",
		EncryptedRoot: encrypted,
		AddChanges:    []ProjectAddChange{{SourcePath: addDir, TargetFolderPath: "source"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(addDir); !os.IsNotExist(err) {
		t.Fatalf("expected added source directory to be removed, stat err = %v", err)
	}
}
