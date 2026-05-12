package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceApplyProjectRenameChange(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
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
		RenameChanges: []ProjectRenameChange{{
			ItemPath: "source/docs",
			NewName:  "papers",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != created.ProjectID || result.AppliedRenames != 1 {
		t.Fatalf("apply result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/papers") {
		t.Fatalf("browser state missing renamed path: %+v", result.BrowserState.Items)
	}
	if browserHasPath(result.BrowserState, "source/docs") {
		t.Fatalf("browser state still has old path: %+v", result.BrowserState.Items)
	}
}

func TestServiceApplyProjectRenameRejectsRoot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
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

	_, err = service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID: created.ProjectID,
		Password:  password,
		RenameChanges: []ProjectRenameChange{{
			ItemPath: "source",
			NewName:  "renamed-source",
		}},
	})
	if err == nil {
		t.Fatal("expected root rename to be rejected")
	}
}

func TestServiceApplyProjectMoveWritesOperationGuide(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
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
		MoveChanges: []ProjectMoveChange{{
			ItemPath:         "source/docs",
			TargetFolderPath: "source/archive",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedMoves != 1 || len(result.ContentOperations) != 1 || result.OperationGuidePath == "" {
		t.Fatalf("apply result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/archive/docs") {
		t.Fatalf("browser state missing moved path: %+v", result.BrowserState.Items)
	}
	data, err := os.ReadFile(result.OperationGuidePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(data)), "move") {
		t.Fatalf("operation guide = %q", data)
	}
}

func TestServiceApplyProjectRemoveDeletesConnectedContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
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
	visiblePath := visiblePathForRealPath(t, ctx, service, created.ProjectID, password, "source/docs/note.txt")

	result, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		RemoveChanges: []ProjectRemoveChange{{
			ItemPath: "source/docs/note.txt",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedRemoves != 1 || len(result.AppliedContentChanges) != 1 || result.OperationGuidePath != "" {
		t.Fatalf("apply result = %+v", result)
	}
	if browserHasPath(result.BrowserState, "source/docs/note.txt") {
		t.Fatalf("browser state still has removed path: %+v", result.BrowserState.Items)
	}
	if _, err := os.Stat(filepath.Join(encrypted, visiblePath)); !os.IsNotExist(err) {
		t.Fatalf("expected encrypted content to be deleted, stat error = %v", err)
	}
}

func TestServiceApplyProjectRemoveConnectedContentMissingBlocksMetadataChange(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
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
	visiblePath := visiblePathForRealPath(t, ctx, service, created.ProjectID, password, "source/note.txt")
	if err := os.Remove(filepath.Join(encrypted, visiblePath)); err != nil {
		t.Fatal(err)
	}

	_, err = service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
		RemoveChanges: []ProjectRemoveChange{{
			ItemPath: "source/note.txt",
		}},
	})
	if err == nil {
		t.Fatal("expected missing connected content to block apply")
	}
	state, err := service.OpenProjectBrowser(ctx, OpenProjectBrowserInput{
		ProjectID: created.ProjectID,
		Password:  password,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !browserHasPath(state, "source/note.txt") {
		t.Fatalf("metadata changed despite missing content: %+v", state.Items)
	}
}

func TestServiceApplyProjectCreateFolderChange(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
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
		CreateFolderChanges: []ProjectCreateFolderChange{{
			TargetFolderPath: "source/docs",
			Name:             "empty",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedCreatedFolders != 1 || result.OperationGuidePath == "" || result.StagedContentPath == "" || len(result.ContentOperations) != 1 {
		t.Fatalf("apply result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/docs/empty") {
		t.Fatalf("browser state missing created folder: %+v", result.BrowserState.Items)
	}
	assertExists(t, filepath.Join(result.StagedContentPath, filepath.FromSlash(result.ContentOperations[0].SourcePath)))
}

func TestServiceApplyProjectCreateFolderUploadsConnectedContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
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
		CreateFolderChanges: []ProjectCreateFolderChange{{
			TargetFolderPath: "source",
			Name:             "empty",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedCreatedFolders != 1 || result.OperationGuidePath != "" || result.StagedContentPath != "" || len(result.AppliedContentChanges) != 1 {
		t.Fatalf("apply result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/empty") {
		t.Fatalf("browser state missing created folder: %+v", result.BrowserState.Items)
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
		t.Fatalf("verify after connected create folder = %+v", verify)
	}
}

func browserHasPath(state ProjectBrowserState, path string) bool {
	for _, item := range state.Items {
		if item.Path == path {
			return true
		}
	}
	return false
}

func visiblePathForRealPath(t *testing.T, ctx context.Context, service Service, projectID, password, realPath string) string {
	t.Helper()
	plan, _, err := service.ReadDatabase(ctx, DatabaseOpen{
		ProjectRef: projectID,
		Password:   password,
	})
	if err != nil {
		t.Fatal(err)
	}
	realPaths, err := projectRealPaths(plan)
	if err != nil {
		t.Fatal(err)
	}
	itemID := ""
	for id, path := range realPaths {
		if path == realPath {
			itemID = id
			break
		}
	}
	if itemID == "" {
		t.Fatalf("real path %q not found", realPath)
	}
	for _, object := range plan.StorageObjects {
		if object.ItemID.String() == itemID {
			return object.VisiblePath
		}
	}
	t.Fatalf("storage object for %q not found", realPath)
	return ""
}
